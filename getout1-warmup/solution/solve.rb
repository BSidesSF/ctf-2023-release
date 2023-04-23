require 'openssl'
require 'base64'
require 'socket'
require 'pp'

require './libronsolve.rb'
require './libgetout'

begin
  s = connect(*get_host_port())
  flag, narrative = use(s, 'ping')

  puts "Fetched narrative: #{ narrative }"
  check_flag(flag, terminate: true)
  exit 0
ensure
  s.close if s
end
