#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "../../lib/libgetout.h"

#define FLAG_FILE "./level1-flag.txt"
#define NARRATIVE_FILE "./level1-narrative.txt"

int main(int argc, char *argv[])
{
  int s = atoi(argv[1]);

  packet_body_t *response = packet_body_create_empty();
  packet_body_add_int(response, 0);
  packet_body_add_file(response, FLAG_FILE, 1);
  packet_body_add_file(response, NARRATIVE_FILE, 0);
  packet_body_send(s, 2, response);
  packet_body_destroy(response);

  return 0;
}
