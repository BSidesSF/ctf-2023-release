#include <stdio.h>
#include <netdb.h>
#include <netinet/in.h>
#include <stdlib.h>
#include <signal.h>
#include <string.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <unistd.h>

#include "server.h"

// #define DEBUG
#define TIMEOUT 60

int g_socket;
static void cleanup(int signal) {
  if(signal) {
    fprintf(stderr, "Caught signal %d\n", signal);
  } else {
    fprintf(stderr, "Cleaning up\n");
  }
  close(g_socket);
  exit(0);
}

void delete_zombies(void)
{
    int status;

    while (waitpid(-1, &status, WNOHANG) > 0) {
    }
}

void start_server(int port, handler_t *handler) {
	struct sockaddr_in servaddr, cli;

  signal(SIGTERM, cleanup);
  signal(SIGINT, cleanup);

	int server = socket(AF_INET, SOCK_STREAM, 0);
	if (server < 0) {
		perror("Couldn't create socket");
		exit(1);
	}

	bzero(&servaddr, sizeof(servaddr));
	servaddr.sin_family = AF_INET;
	servaddr.sin_addr.s_addr = htonl(INADDR_ANY);
	servaddr.sin_port = htons(port);

	if((bind(server, (struct sockaddr*)&servaddr, sizeof(servaddr))) != 0) {
		perror("Bind failed");
		exit(1);
	}

	if((listen(server, 128)) != 0) {
		printf("Listen failed\n");
		exit(1);
	}

  fprintf(stderr, "Listening on port %d..\n", port);
  for(;;) {
    unsigned int len = sizeof(cli);
    int s = accept(server, (struct sockaddr*)&cli, &len);
    if (s < 0) {
      perror("Accept failed");
      exit(1);
    }

    fprintf(stderr, "Accepted connection %d\n", s);

#ifdef DEBUG
    alarm(TIMEOUT);
    close(server);

    fprintf(stderr, "Not forking\n");
    g_socket = s;
    signal(SIGALRM, cleanup);
    handler(g_socket);
    cleanup(0);
    exit(0);
#else
    fprintf(stderr, "Cleaning up zombies...\n");
    signal(SIGCHLD, delete_zombies);

    fprintf(stderr, "Forking\n");
    int child = fork();
    if(!child) {
      // Child
      alarm(TIMEOUT);

      g_socket = s;
      signal(SIGALRM, cleanup);
      handler(g_socket);
      cleanup(0);
      exit(0);
    } else {
      // Server
      close(s);
    }
#endif
  }
}
