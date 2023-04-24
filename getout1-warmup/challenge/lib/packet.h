#ifndef __PACKET_H__
#define __PACKET_H__ value

#include <stdint.h>

#define VERSION 2

typedef enum {
  TYPE_INT = 1,
  TYPE_STRING = 2,
  TYPE_BINARY = 3,
  TYPE_FLOAT = 4,
} arg_type_t;

typedef struct {
  uint32_t version;
  uint32_t crc;
  uint32_t body_length;
  uint32_t arg_count;
} packet_header_t;

typedef struct {
  uint64_t value;
} arg_int_t;

typedef struct {
  char *value;
} arg_string_t;

typedef struct {
  uint32_t length;
  uint8_t *value;
} arg_binary_t;

typedef struct {
  double value;
} arg_float_t;

typedef union {
  arg_int_t i;
  arg_string_t s;
  arg_binary_t b;
  arg_float_t f;
} arg_t;

typedef struct {
  arg_type_t type;
  arg_t value;
} args_t;

typedef struct {
  uint16_t arg_count;
  args_t *args;
} packet_body_t;


void packet_header_print(packet_header_t *header);
packet_header_t *packet_header_read(int s);
void packet_header_destroy(packet_header_t *header);

packet_body_t *packet_body_create_empty();
packet_body_t *packet_body_read(int s, const packet_header_t *header);
void packet_body_add_int(packet_body_t *body, uint64_t value);
void packet_body_add_string(packet_body_t *body, char *value);
void packet_body_add_binary(packet_body_t *body, uint8_t *value, uint32_t length);
void packet_body_add_file(packet_body_t *body, char *filename, uint8_t strip_newline);

/* These return non-zero on failure */
uint8_t packet_read_int_arg(packet_body_t *body, uint32_t arg, uint64_t *value);
uint8_t packet_read_string_arg(packet_body_t *body, uint32_t arg, char **value);
uint8_t packet_read_binary_arg(packet_body_t *body, uint32_t arg, uint8_t **value, uint32_t *length);
uint8_t packet_read_float_arg(packet_body_t *body, uint32_t arg, double *value);

void packet_body_send(int s, int version, packet_body_t *body);
void packet_body_print(packet_body_t *body);
void packet_body_destroy(packet_body_t *body);

#endif /* ifndef __PACKET_H__ */
