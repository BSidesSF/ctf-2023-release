#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>

#define FLAG_FILE "/ctf/level3-flag.txt"
#define NARRATIVE_FILE "/ctf/level3-narrative.txt"

void print_file(char *filename) {
  uint8_t buffer[1024];
  memset(buffer, 0, 1024);
  FILE *f = fopen(filename, "r");

  if(!f) {
    printf("File not found (please report!): %s\n", filename);
  } else {
    fread(buffer, 1, 1023, f);
    fclose(f);

    printf("%s\n", buffer);
  }
}

int main(int argc, char *argv[]) {
  print_file(NARRATIVE_FILE);
  printf("Flag: ");
  print_file(FLAG_FILE);

  return 0;
}
