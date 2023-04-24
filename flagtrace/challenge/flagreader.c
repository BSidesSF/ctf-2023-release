#include <stdlib.h>
#include <stdio.h>
#include <fcntl.h>
#include <unistd.h>

int main (void) {

    int flag_f;
    int key_f;
    ssize_t count;

    void *mem;

    mem = malloc(1024);

    key_f = open("key.bin", O_RDONLY);
    count = read(key_f, mem, 1024);
    close(key_f);

    sprintf(mem, "echo \"Read %ld key bytes\"", count);
    count = system(mem);

    flag_f = open("flag.enc", O_RDONLY);
    count = read(flag_f, mem, 1024);
    close(flag_f);

    sprintf(mem, "echo \"Read %ld flag bytes\"", count);
    count = system(mem);

    return 0;
}
