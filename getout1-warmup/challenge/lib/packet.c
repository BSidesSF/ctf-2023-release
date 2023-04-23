#include <byteswap.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <zlib.h>

#include <arpa/inet.h>
#include <sys/socket.h>

#include "util.h"

#include "packet.h"

#define MAX_FILE_SIZE 2048

void packet_header_print(packet_header_t *header) {
  printf("Version     = 0x%08x\n", header->version);
  printf("CRC         = 0x%08x\n", header->crc);
  printf("Body length = 0x%08x\n", header->body_length);
  printf("Arg count   = 0x%08x\n", header->arg_count);
}

packet_header_t *packet_header_read(int s) {
  packet_header_t *header = (packet_header_t *)malloc(sizeof(packet_header_t));
  uint8_t *data = recv_exactly(s, 16);
  header->version     = bswap_32(((uint32_t*)data)[0]);

  if(header->version != VERSION) {
    fprintf(stderr, "Invalid version: %08x (expected version: %d)\n", header->version, VERSION);
    exit(1);
  }

  header->crc         = bswap_32(((uint32_t*)data)[1]);
  header->body_length = bswap_32(((uint32_t*)data)[2]);
  header->arg_count   = bswap_32(((uint32_t*)data)[3]);
  free(data);

  return header;
}

void packet_header_destroy(packet_header_t *header) {
  free(header);
}

packet_body_t *packet_body_create_empty() {
  packet_body_t *body = malloc(sizeof(packet_body_t));
  body->arg_count = 0;
  body->args = 0;

  return body;
}

void packet_body_add_int(packet_body_t *body, uint64_t value) {
  uint32_t num = body->arg_count;
  body->arg_count += 1;
  body->args = realloc(body->args, sizeof(arg_t) * (body->arg_count + 1));
  body->args[num].type = TYPE_INT;
  body->args[num].value.i.value = value;
}

void packet_body_add_string(packet_body_t *body, char *value) {
  uint32_t num = body->arg_count;
  body->arg_count += 1;
  body->args = realloc(body->args, sizeof(arg_t) * (body->arg_count + 1));
  body->args[num].type = TYPE_STRING;
  body->args[num].value.s.value = (char*) malloc(strlen(value) + 1);
  bzero(body->args[num].value.s.value, strlen(value) + 1);
  strncpy(body->args[num].value.s.value, value, strlen(value));
}

void packet_body_add_binary(packet_body_t *body, uint8_t *value, uint32_t length) {
  uint32_t num = body->arg_count;
  body->arg_count += 1;
  body->args = realloc(body->args, sizeof(arg_t) * (body->arg_count + 1));
  body->args[num].type = TYPE_BINARY;
  body->args[num].value.b.value = (uint8_t*) malloc(length);
  bzero(body->args[num].value.b.value, length);
  memcpy(body->args[num].value.b.value, value, length);
  body->args[num].value.b.length = length;
}

void packet_body_add_file(packet_body_t *body, char *filename, uint8_t strip_newline) {
  uint8_t buffer[MAX_FILE_SIZE];
  memset(buffer, 0, MAX_FILE_SIZE);

  FILE *f = fopen(filename, "r");

  if(!f) {
    packet_body_add_string(body, "I was supposed to add a file, but it wasn't found!");
  } else {
    size_t size = fread(buffer, 1, MAX_FILE_SIZE - 1, f);
    fclose(f);

    if(strip_newline) {
      buffer[strcspn((char*)buffer, "\n")] = '\0';
      packet_body_add_string(body, (char*)buffer);
    } else {
      packet_body_add_binary(body, buffer, size);
    }
  }
}

uint8_t packet_read_int_arg(packet_body_t *body, uint32_t arg, uint64_t *value) {
  if(body->arg_count <= arg) {
    fprintf(stderr, "Missing argument!\n");
    return 1;
  }

  if(body->args[arg].type != TYPE_INT) {
    fprintf(stderr, "Invalid argument type!\n");
    return 1;
  }

  *value = body->args[arg].value.i.value;
  return 0;
}

uint8_t packet_read_string_arg(packet_body_t *body, uint32_t arg, char **value) {
  printf("arg_count: %d\n", body->arg_count);
  if(body->arg_count <= arg) {
    fprintf(stderr, "Missing argument!\n");
    return 1;
  }

  if(body->args[arg].type != TYPE_STRING) {
    fprintf(stderr, "Invalid argument type!\n");
    return 1;
  }

  *value = body->args[arg].value.s.value;
  return 0;
}

uint8_t packet_read_binary_arg(packet_body_t *body, uint32_t arg, uint8_t **value, uint32_t *length) {
  if(body->arg_count <= arg) {
    fprintf(stderr, "Missing argument!\n");
    return 1;
  }

  if(body->args[arg].type != TYPE_BINARY) {
    fprintf(stderr, "Invalid argument type!\n");
    return 1;
  }

  *value = body->args[arg].value.b.value;
  *length = body->args[arg].value.b.length;
  return 0;
}

uint8_t packet_read_float_arg(packet_body_t *body, uint32_t arg, double *value) {
  if(body->arg_count <= arg) {
    fprintf(stderr, "Missing argument!\n");
    return 1;
  }

  if(body->args[arg].type != TYPE_FLOAT) {
    fprintf(stderr, "Invalid argument type!\n");
    return 1;
  }

  *value = body->args[arg].value.f.value;
  return 0;
}


void packet_body_send(int s, int version, packet_body_t *body) {
  uint32_t i;

  uint8_t *body_metadata = (uint8_t*) malloc(8 * body->arg_count);
  uint8_t *body_data = 0;
  uint32_t body_data_length = 0;
  uint32_t length;
  for(i = 0; i < body->arg_count; i++) {
    switch(body->args[i].type) {
      case TYPE_INT:
        length = 8;

        *((uint32_t*)(&body_metadata[8 * i])) = bswap_32(TYPE_INT);
        *((uint32_t*)(&body_metadata[(8 * i) + 4])) = bswap_32(length);

        body_data = realloc(body_data, body_data_length + length);
        *((uint64_t*)(body_data + body_data_length)) = bswap_64(body->args[i].value.i.value);
        body_data_length += length;

        break;

      case TYPE_STRING:
        length = strlen(body->args[i].value.s.value) + 1;
        *((uint32_t*)(&body_metadata[8 * i])) = bswap_32(TYPE_STRING);
        *((uint32_t*)(&body_metadata[(8 * i) + 4])) = bswap_32(length);

        body_data = realloc(body_data, body_data_length + length);
        strncpy((char*)(body_data + body_data_length), body->args[i].value.s.value, length);
        body_data_length += length;
        break;

      case TYPE_BINARY:
        length = body->args[i].value.b.length;
        *((uint32_t*)(&body_metadata[8 * i])) = bswap_32(TYPE_BINARY);
        *((uint32_t*)(&body_metadata[(8 * i) + 4])) = bswap_32(length);

        body_data = realloc(body_data, body_data_length + length);
        memcpy(body_data + body_data_length, body->args[i].value.b.value, length);
        body_data_length += length;
        break;

      case TYPE_FLOAT:
        length = 8;
        *((uint32_t*)(&body_metadata[8 * i])) = bswap_32(TYPE_FLOAT);
        *((uint32_t*)(&body_metadata[(8 * i) + 4])) = bswap_32(length);

        body_data = realloc(body_data, body_data_length + length);
        *((double*)(body_data + body_data_length)) = bswap_64(body->args[i].value.f.value);
        body_data_length += length;
        break;

      default:
        fprintf(stderr, "Unknown type: 0x%08x\n", body->args[i].type);
        exit(1);
    }
  }

  uint32_t crc = crc32(0, body_metadata, 8 * body->arg_count);
  crc = crc32(crc, body_data, body_data_length);

  uint8_t header[16];
  ((uint32_t*)header)[0] = bswap_32(version);
  ((uint32_t*)header)[1] = bswap_32(crc);
  ((uint32_t*)header)[2] = bswap_32(8 * body->arg_count + body_data_length);
  ((uint32_t*)header)[3] = bswap_32(body->arg_count);

  if(send(s, header, 16, 0) < 0) {
    fprintf(stderr, "Failed to send data\n");
    exit(1);
  }

  // printf("Metadata:\n");
  // print_hex(body_metadata, 8 * body->arg_count);
  if(send(s, body_metadata, 8 * body->arg_count, 0) < 0) {
    fprintf(stderr, "Failed to send data\n");
    exit(1);
  }

  // printf("Body:\n");
  // print_hex(body_data, body_data_length);
  if(send(s, body_data, body_data_length, 0) < 0) {
    fprintf(stderr, "Failed to send data\n");
    exit(1);
  }
}

void packet_body_print(packet_body_t *body) {
  uint32_t i;
  uint32_t j;
  for(i = 0; i < body->arg_count; i++) {
    switch(body->args[i].type) {
      case TYPE_INT:
        printf("(int) 0x%016lx\n", body->args[i].value.i.value);
        break;

      case TYPE_STRING:
        printf("(str) %s\n", body->args[i].value.s.value);
        break;

      case TYPE_BINARY:
        printf("(binary) ");

        for(j = 0; j < body->args[i].value.b.length; j++) {
          printf("%02x", body->args[i].value.b.value[j]);
        }

        printf("\n");
        break;
      case TYPE_FLOAT:
        printf("(float) %lf\n", (double)body->args[i].value.f.value);
        break;
    }
  }
}

packet_body_t *packet_body_read(int s, const packet_header_t *header) {
  packet_body_t *body = malloc(sizeof(packet_body_t));

  uint8_t *data = recv_exactly(s, header->body_length);

  // Check the CRC32 before doing anything
  uint32_t crc = crc32(0, data, header->body_length);
  if(crc != header->crc) {
    fprintf(stderr, "Invalid CRC32! Expected %08x, they sent %08x\n", crc, header->crc);
    exit(1);
  }

  uint8_t *data_end = data + header->body_length;

  uint32_t *metadata = (uint32_t*)data;
  uint8_t *bodydata = data + (header->arg_count * 8);
  body->arg_count = 0;
  body->args = (args_t*)malloc(header->arg_count * sizeof(args_t));
  memset(body->args, 0, header->arg_count * sizeof(args_t));

  uint32_t i;
  for(i = 0; i < header->arg_count; i++) {
    if((uint8_t*)&metadata[(i * 2) + 1] > data_end) {
      fprintf(stderr, "Body metadata ran off the end of the packet!\n");
      exit(1);
    }

    uint32_t type = bswap_32(metadata[i * 2]);
    uint32_t extra = bswap_32(metadata[(i * 2) + 1]);

    switch(type) {
      case TYPE_INT:
        if(bodydata + 8 > data_end) {
          fprintf(stderr, "Body data ran off the end of the packet!\n");
          exit(1);
        }

        body->args[i].type = TYPE_INT;
        body->args[i].value.i.value = bswap_64(((uint64_t*)(bodydata))[0]);
        bodydata += 8;
        break;

      case TYPE_STRING:
        body->args[i].type = TYPE_STRING;

        size_t length = strlen((char*)bodydata);
        if(bodydata + length + 1 > data_end) {
          fprintf(stderr, "Body data ran off the end of the packet!\n");
          exit(1);
        }
        body->args[i].value.s.value = (char*)malloc(length + 1);
        memset(body->args[i].value.s.value, 0, length + 1);
        strncpy(body->args[i].value.s.value, (char*)bodydata, length);

        bodydata += length + 1;
        break;

      case TYPE_BINARY:
        body->args[i].type = TYPE_BINARY;

        if(bodydata + extra > data_end) {
          fprintf(stderr, "Body data ran off the end of the packet!\n");
          exit(1);
        }
        body->args[i].value.b.value = (uint8_t*)malloc(extra);
        memset(body->args[i].value.b.value, 0, extra);
        memcpy(body->args[i].value.b.value, bodydata, extra);
        body->args[i].value.b.length = extra;

        bodydata += extra;
        break;

      case TYPE_FLOAT:
        if(bodydata + 8 > data_end) {
          fprintf(stderr, "Body data ran off the end of the packet!\n");
          exit(1);
        }

        body->args[i].type = TYPE_FLOAT;
        body->args[i].value.f.value = bswap_64(((uint64_t*)(bodydata))[0]);
        bodydata += 8;
        break;

      default:
        fprintf(stderr, "Unknown type: 0x%08x\n", type);
        exit(1);
    }

    body->arg_count += 1;
  }

  free(data);

  return body;
}

void packet_body_destroy(packet_body_t *body) {
  uint32_t i;
  for(i = 0; i < body->arg_count; i++) {
    switch(body->args[i].type) {
      case TYPE_STRING:
        free(body->args[i].value.s.value);
        break;

      case TYPE_BINARY:
        free(body->args[i].value.b.value);
        break;

      default:
        // Do nothing
        break;
    }
  }

  free(body);
}
