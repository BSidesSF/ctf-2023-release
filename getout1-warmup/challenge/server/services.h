#ifndef __SERVICES_H__
#define __SERVICES_H__ value

#include <stdint.h>


#define MAX_SERVICES 32
#define MAX_SERVICE_LINE 1024

typedef struct {
  char *name;
  char *executable;
} service_entry_t;

typedef struct {
  uint32_t service_count;
  service_entry_t *services[MAX_SERVICES];
} services_t;

services_t *services_load(char *service_file);
void services_print(services_t *services);
char *services_get_executable(services_t *services, char *name);
void services_destroy(services_t *services);

#endif /* ifndef __SERVICES_H__ */
