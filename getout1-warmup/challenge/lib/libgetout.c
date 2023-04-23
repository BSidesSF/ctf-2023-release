#include <stdarg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#include "packet.h"
#include "util.h"

#include "libgetout.h"

// void test(int s) {
//   packet_body_t *test_out = packet_body_create_empty();
//   packet_body_add_int(test_out, 0x1337);
//   packet_body_print(test_out);
//   packet_body_send(s, 2, test_out);
//   packet_body_destroy(test_out);
// }
//
// void echo(int s) {
//   packet_header_t *header = packet_header_read(s);
//   packet_header_print(header);
//
//   packet_body_t *body = packet_body_read(s, header);
//   packet_body_print(body);
//   packet_body_send(s, 2, body);
//
//   packet_header_destroy(header);
//   packet_body_destroy(body);
// }
//
// void handler(int s) {
//   echo(s);
//   test(s);
// }

void send_simple_response(int s, uint32_t code, char *fmt, ...) {
  char buf[1024];
  bzero(buf, 1024);

  packet_body_t *response = packet_body_create_empty();
  packet_body_add_int(response, code);

  if(fmt) {
    va_list va;
    va_start(va, fmt);
    vsnprintf(buf, 1023, fmt, va);
    va_end(va);

    fprintf(stderr, "Sending code=0x%08x text=%s\n", code, buf);
    packet_body_add_string(response, buf);
  } else {
    fprintf(stderr, "Sending code=0x%08x text=n/a\n", code);
  }

  packet_body_send(s, 2, response);
  packet_body_destroy(response);
}

void send_simple_binary_response(int s, uint32_t code, uint8_t *binary, uint32_t length) {
  fprintf(stderr, "Sending code=0x%08x text=<binary>\n", code);
  packet_body_t *response = packet_body_create_empty();
  packet_body_add_int(response, code);
  packet_body_add_binary(response, binary, length);
  packet_body_send(s, 2, response);
  packet_body_destroy(response);
}
