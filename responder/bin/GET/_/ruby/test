#! /usr/bin/env ruby
require 'rubygems'
require 'bundler/setup'
require 'nats/client'

NATS.start do
  puts "#{ENV['ID']} - responder ready"
  
  NATS.subscribe(ENV['ID'], queue: ENV['ID']) do |msg, reply| 
    puts "Received '#{msg}'"
    response = '{"status":"200","header":{"Content-Type":["text/html; charset=utf-8"]},"body":"Ruby Test"}'
    NATS.publish reply, response
  end
  
end