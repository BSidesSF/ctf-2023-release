#ifndef __LIBGETOUT_H_
#define __LIBGETOUT_H_

#include <stdint.h>

#include "packet.h"
#include "util.h"

void send_simple_response(int s, uint32_t code, char *fmt, ...);
void send_simple_binary_response(int s, uint32_t code, uint8_t *binary, uint32_t length);

#endif /* ifndef __LIBGETOUT_H_ */
