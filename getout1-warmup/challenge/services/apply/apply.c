#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <openssl/aes.h>
#include <openssl/conf.h>
#include <openssl/evp.h>
#include <openssl/err.h>
#include <string.h>

#include "../../lib/libgetout.h"

#ifndef FLAG_FILE
#define FLAG_FILE "./flag.txt"
#endif

#define OPCODE_INSTRUCTIONS 100
#define OPCODE_IDENTIFY 101
#define OPCODE_REGISTER 102
#define OPCODE_CHECK_STATUS 103
#define OPCODE_EXIT 104

#define REGISTER_BUFFER_SIZE 512
#define UUID_SIZE 48

#define TOKEN_FILE "./level3-token.txt"

static uint8_t KEY[] = {0xa6, 0x8b, 0x2a, 0x1c, 0x48, 0x0c, 0xac, 0x26, 0x24, 0x43, 0xab, 0xf9, 0x98, 0xa8, 0x1f, 0x2b};
static uint8_t IV[]  = {0x4a, 0x3b, 0x9f, 0x46, 0x33, 0x0b, 0x76, 0x4f, 0x69, 0x0c, 0x99, 0xc6, 0x62, 0x6f, 0xb9, 0x35};

static int identified = 0;

void handle_instructions(int s) {
  send_simple_response(s, 0,
      "Welcome to the application system!\n"
      "\n"
      "To authorize yourself, please send an IDENTIFY message (opcode 101) with\n"
      "your token\n"
      "\n"
      "Once you've authorized, send a REGISTER message (opcode 102) with five\n"
      "free-form strings, including relevant information about yourself:\n"
      "\n"
      "* Your age and weight\n"
      "* Your diet\n"
      "* Blood type\n"
      "* Ability to toil in mines\n"
      "* Your ideal wine pairing\n"
      "\n"
      "You'll be sent back a verification blob; you can send it to the\n"
      "CHECK_STATUS endpoint (103) to check your status!\n"
  );
}

// Very adapted from https://wiki.openssl.org/index.php/EVP_Symmetric_Encryption_and_Decryption
int encrypt(uint8_t *plaintext, int plaintext_len) {
  uint8_t *buffer = (uint8_t*)malloc(plaintext_len + 16);
  bzero(buffer, plaintext_len + 16);

  EVP_CIPHER_CTX *ctx = EVP_CIPHER_CTX_new();

  int len;

  EVP_EncryptInit_ex(ctx, EVP_aes_128_cbc(), NULL, KEY, IV);
  EVP_EncryptUpdate(ctx, buffer, &len, plaintext, plaintext_len);
  int ciphertext_len = len;
  EVP_EncryptFinal_ex(ctx, buffer + len, &len);
  ciphertext_len += len;
  EVP_CIPHER_CTX_free(ctx);

  memcpy(plaintext, buffer, ciphertext_len);
  free(buffer);

  return ciphertext_len;
}

int decrypt(uint8_t *ciphertext, int ciphertext_len) {
  uint8_t *buffer = (uint8_t*)malloc(ciphertext_len);
  bzero(buffer, ciphertext_len);

  EVP_CIPHER_CTX *ctx = EVP_CIPHER_CTX_new();
  EVP_DecryptInit_ex(ctx, EVP_aes_128_cbc(), NULL, KEY, IV);

  int len;
  if(EVP_DecryptUpdate(ctx, buffer, &len, ciphertext, ciphertext_len) != 1) {
    fprintf(stderr, "EVP_DecryptUpdate failed!\n");
    free(buffer);
    return -1;
  }
  int plaintext_len = len;

  if(EVP_DecryptFinal_ex(ctx, buffer + len, &len) != 1) {
    fprintf(stderr, "EVP_DecryptFinal_ex failed!\n");
    free(buffer);
    return -1;
  }
  plaintext_len += len;

  memcpy(ciphertext, buffer, plaintext_len);
  EVP_CIPHER_CTX_free(ctx);
  free(buffer);

  return plaintext_len;
}

void handle_apply(int s, packet_body_t *body) {
  char validator[] = "uuidgen";
  char buffer[REGISTER_BUFFER_SIZE] = { 0 };
  char uuid[UUID_SIZE] = { 0 };

  if(body->arg_count != 6) {
    send_simple_response(s, ERROR_BAD_REQUEST, "We expect exactly 5 arguments after the opcode (see opcode 100 for instructions)");
    return;
  }

  FILE *uuidgen = popen(validator, "r");
  if(!uuidgen) {
    send_simple_response(s, ERROR_INTERNAL, "Couldn't find uuidgen");
    return;
  }

  fgets(uuid, UUID_SIZE, uuidgen);
  pclose(uuidgen);
  uuid[strcspn(uuid, "\n")] = '\0';

  strncat(buffer + strlen(buffer), body->args[1].value.s.value, REGISTER_BUFFER_SIZE);
  strncat(buffer + strlen(buffer), body->args[2].value.s.value, REGISTER_BUFFER_SIZE);
  strncat(buffer + strlen(buffer), body->args[3].value.s.value, REGISTER_BUFFER_SIZE);
  strncat(buffer + strlen(buffer), body->args[4].value.s.value, REGISTER_BUFFER_SIZE);
  strncat(buffer + strlen(buffer), body->args[5].value.s.value, REGISTER_BUFFER_SIZE);
  strncat(buffer + strlen(buffer), uuid, REGISTER_BUFFER_SIZE);

  size_t length = encrypt((uint8_t*)buffer, strlen(buffer) + 1);
  send_simple_binary_response(s, 0, (uint8_t*)buffer, length);
}

void handle_identify(int s, packet_body_t *body) {
  char *token;

  if(packet_read_string_arg(body, 1, &token)) {
    send_simple_response(s, ERROR_BAD_REQUEST, "Invalid token");
    return;
  }

  FILE *f = fopen(TOKEN_FILE, "r");
  if(!f) {
    send_simple_response(s, ERROR_FILE_NOT_FOUND, "The token file is missing! If this is on the real server, it's a bug, please report");
    return;
  } else {
    char token_buffer[48];
    bzero(token_buffer, 48);
    fgets(token_buffer, 48, f);
    fclose(f);

    token_buffer[strcspn(token_buffer, "\n")] = '\0';

    if(!strcmp(token, token_buffer)) {
      send_simple_response(s, 0, "Token accepted");
      identified = 1;
    } else {
      send_simple_response(s, ERROR_INVALID_TOKEN, "Token rejected");
    }
  }
}

void handle_check_status(int s, packet_body_t *body) {
  uint8_t *buffer;
  uint32_t length;

  if(packet_read_binary_arg(body, 1, &buffer, &length)) {
    send_simple_response(s, ERROR_BAD_REQUEST, "Couldn't read token information");
    return;
  }

  int out_length = decrypt(buffer, length);
  if(out_length < 0) {
    send_simple_response(s, ERROR_DECRYPT_FAILED, "Failed to decrypt your application packet!");
  } else {
    send_simple_binary_response(s, 0, buffer, out_length);
  }
}

void main_loop(int s) {
  for(;;) {
    packet_header_t *header = packet_header_read(s);
    packet_body_t *body = packet_body_read(s, header);

    uint64_t opcode = body->args[0].value.i.value;

    if(opcode == OPCODE_INSTRUCTIONS) {
      fprintf(stderr, "Executing opcode %ld: instructions\n", opcode);
      handle_instructions(s);
    } else if(opcode == OPCODE_IDENTIFY) {
      fprintf(stderr, "Executing opcode %ld: identify\n", opcode);
      handle_identify(s, body);
    } else if(opcode == OPCODE_EXIT) {
      send_simple_response(s, 0, "Goodbye!");
      return;
    } else if(identified) {
      if(opcode == OPCODE_REGISTER) {
        fprintf(stderr, "Executing opcode %ld: apply\n", opcode);
        handle_apply(s, body);
      } else if(opcode == OPCODE_CHECK_STATUS) {
        fprintf(stderr, "Executing opcode %ld: check status\n", opcode);
        handle_check_status(s, body);
      } else {
        send_simple_response(s, ERROR_UNKNOWN_OPCODE, "Unknown opcode: %ld", opcode);
      }
    } else {
      send_simple_response(s, ERROR_UNKNOWN_OPCODE, "Unknown opcode (do you need to identify?)");
    }

    // Intentionally don't destroy them - that makes it easier to get a command into a known address
    /* packet_header_destroy(header); */
    /* packet_body_destroy(body); */
  }
}

int main(int argc, char *argv[]) {
  int s = atoi(argv[1]);

  send_simple_response(s, 0, "Welcome to the Application service, human! Send opcode 100 for instructions");

  main_loop(s);

  return 0;
}
