#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/socket.h>

#include "util.h"

void print_hex(uint8_t *data, uint32_t length) {
  uint32_t i;
  for(i = 0; i < length; i++) {
    printf("%02x", data[i]);
  }
  printf("\n");
}

uint8_t *recv_exactly(int s, uint32_t n) {
  if(n > MAX_RECEIVE) {
    fprintf(stderr, "Trying to receive too much data (requested = %u, max = %u)\n", n, MAX_RECEIVE);
    exit(1);
  }

  uint8_t *buf = malloc(n);
  if(!buf) {
    perror("Failed to allocate memory");
    close(s);
    exit(1);
  }

  uint32_t received = 0;
  while(received < n) {
    ssize_t count = recv(s, buf + received, n - received, 0);
    if(count <= 0) {
      perror("Recv failed");
      exit(1);
    }
    received += count;
  }

  return buf;
}
