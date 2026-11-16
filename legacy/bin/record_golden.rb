#!/usr/bin/env ruby
# Records legacy behaviour into fixtures/golden/temp_tag/*.json ("expected").
require_relative "../lib/boot"
dir = File.expand_path("../../fixtures/golden/temp_tag", __dir__)
Dir[File.join(dir, "*.json")].sort.each do |f|
  fx = JSON.parse(File.read(f))
  fx["expected"] = TempTagIssuer.new(Registration.new(fx["input"])).call
  File.write(f, JSON.pretty_generate(fx) + "\n")
  puts "recorded #{File.basename(f)} -> #{fx["expected"]["status"]}"
end
