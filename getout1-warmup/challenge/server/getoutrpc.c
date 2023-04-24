#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

#include "../lib/libgetout.h"

#include "services.h"
#include "server.h"

#define OPCODE_LIST_SERVICES 0
#define OPCODE_RUN_SERVICE 1

static services_t *services;

void do_run_service(int s, char *service) {
  char *executable = services_get_executable(services, service);
  if(!executable) {
    send_simple_response(s, ERROR_NO_SUCH_SERVICE, "Could not find the requested service");
    return;
  }

  char socket[16];
  sprintf(socket, "%15d", s);

  char *args[]={executable, socket, 0};
  fprintf(stderr, "Executing %s\n", executable);
  execvp(executable, args);

  perror("Couldn't execute the service");
  send_simple_response(s, ERROR_EXEC_FAILED, "Could not execute the requested service");
}

void list_services(int s) {
  packet_body_t *response = packet_body_create_empty();

  uint32_t i;
  for(i = 0; i < services->service_count; i++) {
    packet_body_add_string(response, services->services[i]->name);
  }
  packet_body_send(s, 2, response);
  packet_body_destroy(response);
}

void run_service(int s, packet_body_t *body) {
  if(body->arg_count != 2) {
    send_simple_response(s, ERROR_BAD_REQUEST, "Wrong number of arguments in run-service request");
  } else {
    char *service_name;
    if(packet_read_string_arg(body, 1, &service_name)) {
      send_simple_response(s, ERROR_BAD_REQUEST, "Wrong argument type");
    } else {
      do_run_service(s, service_name);
    }
  }
}

void service_executor(int s) {
  for(;;) {
    packet_header_t *header = packet_header_read(s);
    packet_body_t *body = packet_body_read(s, header);

    uint64_t opcode;

    if(packet_read_int_arg(body, 0, &opcode)) {
      send_simple_response(s, ERROR_BAD_REQUEST, "Wrong argument type");
      return;
    }

    if(opcode == OPCODE_LIST_SERVICES) {
      fprintf(stderr, "Executing opcode %ld: list services\n", opcode);
      list_services(s);
    } else if(opcode == OPCODE_RUN_SERVICE) {
      fprintf(stderr, "Executing opcode %ld: run service\n", opcode);
      run_service(s, body);
    } else {
      send_simple_response(s, ERROR_UNKNOWN_OPCODE, "Unknown opcode: %d", opcode);
      return;
    }

    packet_header_destroy(header);
    packet_body_destroy(body);
  }
}

int main(int argc, char *argv[]) {
  if(argc < 2) {
    printf("You forgot args!\n");
    exit(1);
  }

  services = services_load(argc > 2 ? argv[2] : "/ctf/rpcservices");
  services_print(services);

  start_server(atol(argv[1]), &service_executor);

  services_destroy(services);

  return 0;
}
