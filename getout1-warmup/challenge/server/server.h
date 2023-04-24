#ifndef __SERVER_H__
#define __SERVER_H__

typedef void (handler_t)(int s);

void start_server(int port, handler_t *handler);

#endif
