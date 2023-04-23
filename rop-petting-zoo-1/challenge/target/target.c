#include <signal.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>
#include <sys/mman.h>

#define ALLOC_AMOUNT 2048
#define BASE ((void*)0x13370000)

#define FLAG1 "CTF{getting-started-on-rop}"
#define FLAG2 "CTF{getting-more-good-at-rop}"
#define FLAG3 "CTF{good-ropping-or-cheating}"

char *flag_filename = NULL;
void *writable_memory;

void exitfunc() {
  if(flag_filename) {
    unlink(flag_filename);
    munmap(flag_filename, L_tmpnam + 1);
    flag_filename = NULL;
  }
}

void sighandler(int no) {
  exit(no);
}

typedef struct {
  char *code;
  uint8_t length;
  char *instruction;
  char *description;
  uint8_t prompt_for_integer;
} entry_t;

typedef struct {
  void *function;
  char *name;
  char *description;
  uint32_t argcount;
} function_entry_t;

void usage(char *name) {
  fprintf(stderr, "Usage: %s <check|run> [code]\n", name);
  exit(0);
}

typedef enum {
  CHECK,
  RUN,
} run_mode_t;

void *get_writable_memory() {
  return writable_memory;
}

char *get_etc_passwd() {
  return "/etc/passwd";
}

char *get_hello_world() {
  return "Hello world!";
}

char *get_letter_r() {
  return "r";
}

void print_flag() {
  printf("%s\n", FLAG1);
}

char *return_flag() {
  return FLAG2;
}

char *write_flag_to_file() {
  if(flag_filename) {
    unlink(flag_filename);
    munmap(flag_filename, L_tmpnam + 1);
    flag_filename = NULL;
  }

  flag_filename = mmap((void*)(uint64_t)(rand() & 0xFFFFFF00), L_tmpnam + 1, PROT_READ | PROT_WRITE, MAP_ANONYMOUS | MAP_PRIVATE, 0, 0);
  tmpnam(flag_filename);

  FILE *f = fopen(flag_filename, "w");
  fputs(FLAG3, f);
  fclose(f);

  return flag_filename;
}

void main(int argc, char *argv[]) {
  alarm(30);
  srand(time(NULL));

  if(argc < 2) {
    usage(argv[0]);
  }

  run_mode_t mode;

  if(!strcmp(argv[1], "check")) {
    mode = CHECK;
  } else if(!strcmp(argv[1], "run")) {
    mode = RUN;

    if(argc < 3) {
      usage(argv[0]);
    }
  } else {
    usage(argv[0]);
  }

  entry_t entries[] = {
    { "\x5F\xC3",         2, "pop rdi / ret",     "Remove the top of the stack and put it in rdi", 1},
    { "\x5E\xC3",         2, "pop rsi / ret" ,    "Remove the top of the stack and put it in rsi", 1},
    { "\x5A\xC3",         2, "pop rdx / ret" ,    "Remove the top of the stack and put it in rdx", 1 },
    { "\x59\xC3",         2, "pop rcx / ret" ,    "Remove the top of the stack and put it in rcx", 1 },
    { "\x48\x89\xC7\xC3", 4, "mov rdi, rax / ret", "Set rdi to the most recent return value (rax)", 0 },
    { "\x48\x89\xC6\xC3", 4, "mov rsi, rax / ret", "Set rsi to the most recent return value (rax)", 0 },
    { "\x48\x89\xC2\xC3", 4, "mov rdx, rax / ret", "Set rdx to the most recent return value (rax)", 0 },
    { "\x48\x89\xC1\xC3", 4, "mov rcx, rax / ret", "Set rcx to the most recent return value (rax)", 0 },
    { "\x48\xFF\xC7\xC3", 4, "inc rdi / ret" ,    "Increment rdi by 1", 0 },
    { "\x48\xFF\xC6\xC3", 4, "inc rsi / ret" ,    "Increment rsi by 1", 0 },
    { "\x48\xFF\xC2\xC3", 4, "inc rdx / ret" ,    "Increment rdx by 1", 0 },
    { "\x48\xFF\xC1\xC3", 4, "inc rcx / ret" ,    "Increment rcx by 1", 0 },
    { 0, 0, 0, 0 }
  };

  function_entry_t functions[] = {
    {
      sleep,
      "sleep(seconds)",
      "Sleep for <rdi> seconds",
      1
    },
    {
      fopen,
      "fopen(path, mode)",
      "Open <rdi> with the mode <rsi>",
      2
    },
    {
      fgets,
      "fgets(buf, size, stream)",
      "Read <rsi> bytes from the open file <rdx> into the buffer <rdi>",
      3
    },
    {
      puts,
      "puts(s)",
      "Print the string <rdi>",
      1
    },
    {
      exit,
      "exit(code)",
      "Exit with the code <rdi>",
      1
    },
    {
      get_writable_memory,
      "get_writable_memory()",
      "Return a pointer to a 1024-byte buffer (will always return the same buffer)",
      0
    },
    {
      get_hello_world,
      "get_hello_world()",
      "Return a pointer to the string \\\"Hello world!\\\"",
      0
    },
    {
      get_etc_passwd,
      "get_etc_passwd()",
      "Return a pointer to the string \\\"/etc/passwd\\\"",
      0
    },
    {
      get_letter_r,
      "get_letter_r()",
      "Returns a pointer to the string \\\"r\\\"",
      0
    },
    { 0, 0, 0 }
  };

  function_entry_t targets[] = {
    {
      print_flag,
      "print_flag()",
      "Prints the level 1 flag to stdout (which you can see)",
      0
    },
    {
      return_flag,
      "return_flag()",
      "Returns a pointer to the level 2 flag, which can be printed by a function such as puts()",
      0
    },
    {
      write_flag_to_file,
      "write_flag_to_file()",
      "Writes the level 3 flag to a random file, which can be accessed with functions such as fopen() / fgets() / puts()",
      0
    },

    { 0, 0, 0 }
  };

  uint8_t *code_buffer = mmap(BASE, ALLOC_AMOUNT, PROT_READ | PROT_WRITE | PROT_EXEC, MAP_ANONYMOUS | MAP_PRIVATE, 0, 0);
  uint8_t *ptr_code_buff = code_buffer;

  if(mode == CHECK) {
    printf("[\n");
  }

  /* We always have to build the gadgets array */
  int i;
  for(i = 0; entries[i].code; i++) {
    memcpy(ptr_code_buff, entries[i].code, entries[i].length);
    if(mode == CHECK) {
      printf("  {\n");
      printf("    \"name\": \"%s\",\n", entries[i].instruction);
      printf("    \"address\": %lu,\n", (uint64_t)ptr_code_buff);
      printf("    \"description\": \"%s\",\n", entries[i].description);
      printf("    \"type\": \"gadgets\",\n");
      printf("    \"prompt_for_integer\": %d\n", entries[i].prompt_for_integer);
      printf("  },\n");
    }
    ptr_code_buff += entries[i].length;
  }

  /* We only need to print the functions in CHECK mode */
  if(mode == CHECK) {
    for(i = 0; functions[i].function; i++) {
      printf("  {\n");
      printf("    \"name\": \"%s\",\n", functions[i].name);
      printf("    \"address\": %lu,\n", (uint64_t) functions[i].function);
      printf("    \"description\": \"%s\",\n", functions[i].description);
      printf("    \"type\": \"functions\",\n");
      printf("    \"argcount\": %d,\n", functions[i].argcount);
      printf("    \"prompt_for_integer\": 0\n");
      printf("  },\n");
    }

    for(i = 0; targets[i].function; i++) {
      printf("  {\n");
      printf("    \"name\": \"%s\",\n", targets[i].name);
      printf("    \"address\": %lu,\n", (uint64_t) targets[i].function);
      printf("    \"description\": \"%s\",\n", targets[i].description);
      printf("    \"type\": \"target-function\",\n");
      printf("    \"argcount\": %d,\n", targets[i].argcount);
      printf("    \"prompt_for_integer\": 0\n");
      printf("  }\n");

      /* Don't print a comma on the last one (JSON!!) */
      if(targets[i + 1].function) {
        printf(",\n");
      }
    }

    /* Terminate */
    printf("\n]\n");
    exit(0);
  }

  // Set up the writable memory
  writable_memory = malloc(1024);
  bzero(writable_memory, 1024);

  uint8_t *return_address = ((uint8_t*)__builtin_frame_address(0)) + 8;
  for(i = 0; i < strlen(argv[2]); i += 2) {
    sscanf(argv[2] + i, "%2hhx", return_address);
    return_address++;
  }

  // Make sure the next thing is nothing
  bzero(return_address, 16);

  // This will delete temp files
  atexit(exitfunc);
  signal(SIGSEGV, sighandler);
  signal(SIGILL, sighandler);
  signal(SIGTERM, sighandler);
  signal(SIGALRM, sighandler);

  __asm__("int $3");
}
