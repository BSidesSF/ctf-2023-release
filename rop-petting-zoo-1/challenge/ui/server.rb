require 'json'
require 'sinatra'
require 'pp'
require 'commonmarker'
require 'singlogger'

# This is where js/css/etc goes
set :public_folder, 'public'

::SingLogger.set_level_from_string(level: ENV['log_level'] || 'debug')
LOGGER = ::SingLogger.instance()

# Ideally, we set all these in the Dockerfile
set :bind, ENV['HOST'] || '0.0.0.0'
set :port, ENV['PORT'] || '8080'
MANDRAKE_PATH = ENV['MANDRAKE'] || '../mandrake'
TARGET_PATH = ENV['TARGET'] || '../target/target'

LOGGER.info("Checking for required binaries...")
if File.exists?(MANDRAKE_PATH)
  LOGGER.info("* Found `mandrake` binary: #{ MANDRAKE_PATH }")
else
  LOGGER.fatal("* Couldn't find `mandrake` binary #{ MANDRAKE_PATH } - use the `MANDRAKE` env var to change")
  exit(1)
end

if File.exists?(TARGET_PATH)
  LOGGER.info("* Found `target` binary: #{ TARGET_PATH }")
else
  LOGGER.fatal("* Couldn't find `target` binary #{ TARGET_PATH } - use the `TARGET` env var to change")
  exit(1)
end

# Get addresses
LOGGER.info("Running #{ TARGET_PATH } twice to make sure ASLR isn't going to mess us up...")
ADDRESSES = JSON::parse(`#{ TARGET_PATH } check`)
TEST_ADDRESSES = JSON::parse(`#{ TARGET_PATH } check`)

if TEST_ADDRESSES != ADDRESSES
  LOGGER.fatal("* Running #{ TARGET_PATH } twice output different addresses. Make sure ASLR is off!")
  LOGGER.fatal("* This could work: `echo 0 | sudo tee /proc/sys/kernel/randomize_va_space`")
  exit(1)
else
  LOGGER.info("ASLR looks good!")
end

ADDRESSES_FOR_DECODING = ADDRESSES.map do |a|
  [a['address'], {
    name: a['name'],
    argcount: a['argcount'],
  }]
end.to_h

ARGS = ['rdi', 'rsi', 'rdx', 'rcx']

LEVELS = [
  {
    name: 'tutorial1',
    title: 'Tutorial 1: Basics',
    text: CommonMarker.render_html(File.read('levels/tutorial1.md'))
  },
  {
    name: 'level1',
    title: 'Level 1: Basics',
    text: CommonMarker.render_html(File.read('levels/level1.md'))
  },
  {
    name: 'tutorial2',
    title: 'Tutorial 2: Return values',
    text: CommonMarker.render_html(File.read('levels/tutorial2.md'))
  },
  {
    name: 'level2',
    title: 'Level 2: Using a Return Value',
    text: CommonMarker.render_html(File.read('levels/level2.md'))
  },
  {
    name: 'tutorial3',
    title: 'Tutorial 3: Files',
    text: CommonMarker.render_html(File.read('levels/tutorial3.md'))
  },
  {
    name: 'level3',
    title: 'Level 3: Reading a file',
    text: CommonMarker.render_html(File.read('levels/level3.md'))
  },
]

get '/' do
  erb :index, :locals => { levels: LEVELS }
  # send_file File.join(settings.public_folder, 'index.html')
end

get '/gadgets' do
  content_type 'application/json'

  # Since JavaScript is JavaScript, generate the strings here
  gadgets = JSON::parse(`\"#{ TARGET_PATH }\" check`).map do |gadget|
    gadget['hex'] = [gadget['address']].pack('Q').unpack('H*').pop
    gadget['address'] = "0x%x" % gadget['address']

    gadget
  end.to_json()

  return gadgets
end

post '/execute' do
  content_type 'application/json'

  # Parse the data
  begin
    data = request.body.read
    data = JSON.parse(data)
  rescue JSON::ParserError => e
    LOGGER.error("Couldn't parse JSON from client: #{ e.to_s }")
    return { 'error' => 'Bad JSON in body' }.to_json
  end


  # Ensure that the code is sensible
  begin
    LOGGER.debug("Code: #{data['code']}")

    if data['code'] !~ /^([a-fA-F0-9]{16})*$/
      LOGGER.debug("User used invalid hex in their code!")
      return { 'error' => 'Invalid code' }.to_json
    end

    out = JSON::parse(`"#{ MANDRAKE_PATH }" -o JSON --minimum-viable-string 0 --max-instructions 65536 elf "#{ TARGET_PATH }" run "#{ data['code'] }"`)

    # Clean up the history to just what we want the player to see
    out['history'] = out['history'].select do |s|
      rip = s['rip']['value']

      # Allow stuff in our 0x13370000 block, or in our known addresses
      rip & 0x00000000FFFF0000 == 0x13370000 || ADDRESSES_FOR_DECODING[rip]
    end.map do |s|
      entry = {}
      rip = s['rip']

      if rip['value'] & 0x00000000FFFF0000 == 0x13370000
        entry['instruction'] = rip['as_instruction']
      else
        addr_info = ADDRESSES_FOR_DECODING[rip['value']]
        args = []

        0.upto(addr_info[:argcount] - 1) do |i|
          arg = s[ARGS[i]]

          if arg['as_string']
            if arg['as_string'] =~ /ctf/i
              args << "#{ ARGS[i] } = 0x#{arg['value'].to_s(16)} (\"<censored the flag>\")"
            else
              args << "#{ ARGS[i] } = 0x#{arg['value'].to_s(16)} (\"#{ arg['as_string'] }\")"
            end
          else
            args << "#{ ARGS[i] } = 0x#{arg['value'].to_s(16)}"
          end
        end

        entry['instruction'] = "#{ addr_info[:name].gsub(/\(.*/, '') }(#{ args.join(', ') })"
      end

      entry
    end

    out.to_json()
  rescue Exception => e
    LOGGER.fatal("Something went wrong: #{ e }")
    $stderr.puts(e.backtrace)

    return { 'error' => "Oops, something went wrong! Please report (unless you were intentionally messing around)!" }.to_json
  end
end
