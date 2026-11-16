#!/usr/bin/env ruby
require_relative "../lib/boot"
$stdout.sync = true
$stdin.each_line do |l|
  next if l.strip.empty?
  puts JSON.generate(TempTagIssuer.new(Registration.new(JSON.parse(l))).call)
end
