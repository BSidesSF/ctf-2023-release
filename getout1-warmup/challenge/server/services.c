#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#include "services.h"

services_t *services_load(char *service_file) {
  FILE *f = fopen(service_file, "r");
  if(!f) {
    fprintf(stderr, "Couldn't find load file: %s\n", service_file);
    return NULL;
  }

  services_t *services = (services_t*) malloc(sizeof(services_t));
  memset(services, 0, sizeof(services_t));

  for(;;) {
    char buffer[MAX_SERVICE_LINE];
    if(!fgets(buffer, MAX_SERVICE_LINE, f)) {
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

    service_entry_t *entry = (service_entry_t*) malloc(sizeof(service_entry_t));
    memset(entry, 0, sizeof(service_entry_t));

    entry->name = malloc(strlen(buffer) + 1);
    bzero(entry->name, strlen(buffer) + 1);
    strncpy(entry->name, buffer, strlen(buffer));

    entry->executable = malloc(strlen(space) + 1);
    bzero(entry->executable, strlen(space) + 1);
    strncpy(entry->executable, space, strlen(space));

    services->services[services->service_count++] = entry;
  }

  return services;
}

void services_print(services_t *services) {
  if(!services || services->service_count == 0) {
    printf("No services loaded\n");
    return;
  }

  uint32_t i;
  for(i = 0; i < services->service_count; i++) {
    printf("%s executes %s\n", services->services[i]->name, services->services[i]->executable);
  }
}

char *services_get_executable(services_t *services, char *name) {
  if(!services) {
    fprintf(stderr, "No services loaded, defaulting to insecure mode\n");
    return name;
  }

  uint32_t i;
  for(i = 0; i < services->service_count; i++) {
    if(!strcmp(services->services[i]->name, name)) {
      return services->services[i]->executable;
    }
  }

  return 0;
}

void services_destroy(services_t *services) {
  if(!services || services->service_count == 0) {
    printf("No services loaded\n");
    return;
  }

  uint32_t i;
  for(i = 0; i < services->service_count; i++) {
    free(services->services[i]->name);
    free(services->services[i]->executable);
    free(services->services[i]);
  }
  free(services);
}
