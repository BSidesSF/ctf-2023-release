#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/types.h>
#include <pwd.h>

#include "../../lib/libgetout.h"

#define FLAG_FILE "./level2-flag.txt"
#define NARRATIVE_FILE "./level2-narrative.txt"
#define USERS_FILE "./users"

uint8_t check_creds(char *username, char *password) {
  printf("Checking credentials: %s / %s\n", username, password);
  sleep(1);

  if(!strcmp(username, ":testuser:")) {
    // Test user

    char *uid = strchr(password, ':');
    if(!uid) {
      return 0;
    }

    *uid = '\0';
    uid++;

    char *gid = strchr(uid, ':');
    if(!gid) {
      return 0;
    }
    *gid = '\0';
    gid++;

    struct passwd *userinfo = getpwnam(password);
    if(!userinfo) {
      return 0;
    }

    if(userinfo->pw_uid != atoi(uid)) {
      return 0;
    }

    if(!atoi(gid)) {
      return 0;
    }


    return 1;
  } else {
    // Normal user
    FILE *users = fopen(USERS_FILE, "r");

    // In reality, this will happen since we don't actually have a users file
    if(!users) {
      return 0;
    }

    for(;;) {
      char buffer[1024];
      if(!fgets(buffer, 1023, users)) {
        break;
      }

      // Ignore comments
      if(buffer[0] == '#') {
        continue;
      }

      // Remove the newline
      buffer[strcspn(buffer, "\n")] = '\0';

      char *space = strchr(buffer, ' ');
      if(!space) {
        continue;
      }

      *space = '\0';
      space++;

      if(!strcmp(buffer, username) && !strcmp(space, password)) {
        return 1;
      }
    }
    return 0;
  }
}

int main(int argc, char *argv[]) {
  int s = atoi(argv[1]);

  send_simple_response(s, 0, "Please authenticate with the credentials you were issued");
  for(;;) {
    packet_header_t *header = packet_header_read(s);
    packet_body_t *body = packet_body_read(s, header);

    char *username;
    if(packet_read_string_arg(body, 0, &username)) {
      send_simple_response(s, ERROR_BAD_REQUEST, "Couldn't read username");
      exit(1);
    }

    char *password;
    if(packet_read_string_arg(body, 1, &password)) {
      send_simple_response(s, ERROR_BAD_REQUEST, "Couldn't read password for %s", username);
      exit(1);
    }

    if(check_creds(username, password)) {
      packet_body_t *response = packet_body_create_empty();
      packet_body_add_int(response, 0);
      packet_body_add_file(response, FLAG_FILE, 1);
      packet_body_add_file(response, NARRATIVE_FILE, 0);
      packet_body_send(s, 2, response);
      packet_body_destroy(response);
      //exit(0);
    } else {
      send_simple_response(s, ERROR_INVALID_LOGIN, "Invalid login: %s / %s", username, password);
    }

    packet_header_destroy(header);
    packet_body_destroy(body);
  }

  return 0;
}
