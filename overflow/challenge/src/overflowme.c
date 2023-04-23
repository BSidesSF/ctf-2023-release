/* Providing this for reference */

#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <errno.h>

int main(int argc, char *argv[]) {
	int value = 5;
	char buffer_one[8], buffer_two[8];

	strcpy(buffer_one, "one"); /* put "one" into buffer_one */
	strcpy(buffer_two, "two"); /* put "two" into buffer_two */

	printf("[BEFORE] buffer_two is at %p and contains \'%s\'\n", buffer_two, buffer_two);
	printf("[BEFORE] buffer_one is at %p and contains \'%s\'\n", buffer_one, buffer_one);
	printf("[BEFORE] value is at %p and is %d (0x%08x)\n", &value, value, value);

	printf("\n[STRCPY] copying %d bytes into buffer_two\n\n",  strlen(argv[1]));
	strcpy(buffer_two, argv[1]); /* copy first argument into buffer_two */

	printf("[AFTER] buffer_two is at %p and contains \'%s\'\n", buffer_two, buffer_two);
	printf("[AFTER] buffer_one is at %p and contains \'%s\'\n", buffer_one, buffer_one);
	printf("[AFTER] value is at %p and is %d (0x%08x)\n", &value, value, value);

  /* Added for the CTF! */
  if(!strcmp(buffer_one, "hacked")) {
    char buffer[64];
    FILE *f = fopen("/home/ctf/flag.txt", "r");

    if(!f) {
      printf("\n\nFailed to open flag.txt: %s\n", strerror(errno));

      exit(1);
    }

    fgets(buffer, 63, f);
    printf("\n\nCongratulations! %s\n", buffer);
    exit(0);
  } else {
    printf("\n\nPlease set buffer_one to \"hacked\"!\n");
  }
}
