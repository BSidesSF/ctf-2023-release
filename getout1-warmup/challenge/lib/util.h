#ifndef __UTIL__H__
#define __UTIL__H__ value

#include <stdint.h>

#define MAX_RECEIVE 32768

#define ERROR_BAD_REQUEST     31000
#define ERROR_NO_SUCH_SERVICE 31001
#define ERROR_UNKNOWN_OPCODE  31002
#define ERROR_EXEC_FAILED     31003
#define ERROR_FILE_NOT_FOUND  31004
#define ERROR_INTERNAL        31005
#define ERROR_INVALID_LOGIN   31006
#define ERROR_INVALID_TOKEN   31007
#define ERROR_DECRYPT_FAILED  31008

void print_hex(uint8_t *data, uint32_t length);
uint8_t *recv_exactly(int s, uint32_t n);

#endif /* ifndef __UTIL__H__ */
