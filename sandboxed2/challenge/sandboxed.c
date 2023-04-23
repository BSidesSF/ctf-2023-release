#include <stdio.h>
#include <unistd.h>
#include <fcntl.h>
#include <errno.h>
#include <string.h>
#include <sys/mman.h>

#include <seccomp.h>

typedef struct _denylist_entry {
  size_t len;
  char *pattern;
} denylist_entry;

void init_sandbox();
char *alloc_shellcode(size_t len);
void read_shellcode(int fd, char *buf, size_t len);
void exec_shellcode(char *buf);
void fixup_shellcode(char *ptr, size_t len);
int buf_has_prefix(const char *buf, const char *pfx, size_t len);

#define SC_SIZE 1024

denylist_entry denylist[] = {
  {.len=2, .pattern="\x0f\x05"},
  {.len=2, .pattern="\x0f\x34"},
  {.len=2, .pattern="\xcd\x80"},
  {.len=0, .pattern=""},
};

int main(int argc, char **argv) {
		setvbuf(stdout, NULL, _IONBF, 0);
		char *buf = alloc_shellcode(SC_SIZE);
		printf("I'd love exactly %d bytes of shellcode.\n", SC_SIZE);
		read_shellcode(STDIN_FILENO, buf, SC_SIZE);
		fixup_shellcode(buf, SC_SIZE);
		printf("Initializing sandbox... ");
		init_sandbox();
		printf("done!\n");
		exec_shellcode(buf);
		return 0;
}

char *alloc_shellcode(size_t len) {
		void *rv = mmap(NULL, len, PROT_READ|PROT_WRITE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0);
		if (!rv || rv == MAP_FAILED) {
				printf("mmap failed: %s\n", strerror(errno));
				_exit(1);
		}
		return (char *)rv;
}

void fixup_shellcode(char *ptr, size_t len) {
  for (size_t i=0; i<len; i++) {
    denylist_entry *e = &denylist[0];
    size_t left = len - i;
    while (e->len) {
      if (e->len <= left) {
        if (buf_has_prefix(&ptr[i], e->pattern, e->len)) {
          _exit(2);
        }
      }
      e++;
    }
  }
  if (mprotect(ptr, len, PROT_READ|PROT_EXEC) != 0) {
    _exit(1);
  }
}

int buf_has_prefix(const char *buf, const char *pfx, size_t len) {
  if (memcmp((const void *)buf, (const void *)pfx, len) == 0) {
    return 1;
  }
  return 0;
}

void read_shellcode(int fd, char *buf, size_t len) {
		size_t done = 0;
		while (done < len) {
				size_t l = read(fd, (void *)(buf+done), len-done);
				if (l <= 0) {
						// read error!
						_exit(2);
				}
				done += l;
		}
}

void exec_shellcode(char *buf) {
		void (*func)() = (void(*)())buf;
		func();
}

void init_sandbox() {
		scmp_filter_ctx sandbox_ctx = seccomp_init(SCMP_ACT_KILL);
		if (!sandbox_ctx) {
				_exit(1);
		}
		int rv;
#define MUST_ADD(action, syscall, arg_cnt, ...) if ((rv = (seccomp_rule_add(sandbox_ctx, action, SCMP_SYS(syscall), arg_cnt, ##__VA_ARGS__))) != 0) _exit(1)

		/* Rules here */
		MUST_ADD(SCMP_ACT_ALLOW, exit, 0);
		MUST_ADD(SCMP_ACT_ALLOW, exit_group, 0);
		MUST_ADD(SCMP_ACT_ALLOW, read, 0);
		MUST_ADD(SCMP_ACT_ALLOW, write, 0);
		MUST_ADD(SCMP_ACT_ALLOW, readv, 0);
		MUST_ADD(SCMP_ACT_ALLOW, writev, 0);
		/* munmap for freeing memory w/musl */
		MUST_ADD(SCMP_ACT_ERRNO(0), munmap, 0);
		// allow open for reading
		MUST_ADD(SCMP_ACT_ALLOW, open, 1,
						SCMP_A1_64(SCMP_CMP_EQ, O_RDONLY));

		if ((rv = seccomp_load(sandbox_ctx)) != 0) {
#ifdef DEBUG
				fprintf(stderr, "Error loading seccomp: %d\n", rv);
				fflush(stderr);
#endif
				_exit(1);
		}
		seccomp_release(sandbox_ctx);
}
