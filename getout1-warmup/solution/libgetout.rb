require 'openssl'
require 'base64'
require 'yaml'
require 'socket'
require 'pp'
require 'zlib'

TYPE_INT = 1
TYPE_STRING = 2
TYPE_BINARY = 3
TYPE_FLOAT = 4

def get_header(version, crc, body_length, arg_count)
  return [version, crc, body_length, arg_count].pack('NNNN')
end

def get_body(args)
  # puts "out:"
  # pp args
  # puts

  metadata_buffer = ''
  data_buffer = ''

  args.each do |a|
    if a[:type] == :integer
      metadata_buffer += [TYPE_INT, 0x1234].pack('NN')
      data_buffer += [a[:value]].pack('Q>')
    elsif a[:type] == :string
      metadata_buffer += [TYPE_STRING, 0x4321].pack('NN')

      if a[:nonull]
        data_buffer += a[:value]
      else
        data_buffer += [a[:value]].pack('Z*')
      end

    elsif a[:type] == :binary
      metadata_buffer += [TYPE_BINARY, a[:value].length].pack('NN')
      data_buffer += a[:value]
    elsif a[:type] == :float
      metadata_buffer += [TYPE_FLOAT, 0x1234].pack('NN')
      data_buffer += [a[:value]].pack('Q>')
    else
      throw "Unknown type: #{ a[:type] }"
    end
  end

  return metadata_buffer + data_buffer
end

def create_message(args)
  body = get_body(args)

  header = get_header(2, Zlib.crc32(body), body.length, args.length)

  return header + body
end

def read_message(s)
  header = s.recv(16)
  if header.nil? || header.empty?
    $stderr.puts "Connection closed?"
    exit 1
  end

  if header.length != 16
    $stderr.puts "Didn't receive full header?"
    exit 1
  end

  version, crc, body_length, arg_count = header.unpack('NNNN')

  body = ''
  loop do
    recvd = s.recv(body_length - body.length)
    if recvd.nil? || recvd.empty?
      $stderr.puts "Socket closed?"
      exit 1
    end

    body += recvd
    if body.length >= body_length
      break
    end
  end

  # Check CRC32
  real_crc = Zlib.crc32(body)

  if real_crc != crc
    $stderr.puts "crc didn't match!"
    puts "real = %08x :: packet had %08x" % [real_crc, crc]
    exit 1
  end

  # puts "Version = #{ version }"
  # puts "CRC = #{ crc }"
  # puts "Body length = #{ body_length }"
  # puts "Arg count = #{ arg_count }"
  # puts "Raw metadata/body: #{ body.unpack('H*') }"
  metadata, body = body.unpack("a#{ arg_count * 8 }a*")
  metadata = metadata.unpack('N*')
  args = []


  0.upto(arg_count - 1) do |i|
    type = metadata[i * 2]
    extra = metadata[1 + (i * 2)]

    case type
    when TYPE_INT
      value, body = body.unpack('Q>a*')
      args << {
        type: :integer,
        value: value,
      }

    when TYPE_STRING
      value, body = body.unpack('Z*a*')
      args << {
        type: :string,
        value: value,
      }

    when TYPE_BINARY
      value, body = body.unpack("a#{extra}a*")
      args << {
        type: :binary,
        value: value,
      }

    when TYPE_FLOAT
      throw 'TODO'
    else
      throw "Unknown type"
    end
  end

  # puts "in:"
  # pp args
  # puts

  return args
end

def connect(host, port)
  s = TCPSocket.new(host, port)

  s.write(create_message([
    {
      type: :integer,
      value: 0,
    }
  ]))
  out = read_message(s)

  puts "Connected to RPC service! RPC services available:"
  puts out.map { |o| "* #{ o[:value] }" }.join("\n")
  puts

  return s
end

def use(s, service)
  s.write(create_message([
    {
      type: :integer,
      value: 1,
    },
    {
      type: :string,
      value: service,
    },
  ]))

  out = read_message(s)
  if out[0][:value] != 0
    throw "Something went wrong while USE'ing the service: #{ out[1][:value] }"
  else
    return out[1][:value], out[2] ? out[2][:value] : nil
  end
end
