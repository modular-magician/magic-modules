require 'vcr'

VCR.configure do |c|
  c.hook_into :webmock
  c.cassette_library_dir = 'asdf'
  c.allow_http_connections_when_no_cassette = true
end