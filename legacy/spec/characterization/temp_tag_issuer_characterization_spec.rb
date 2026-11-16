require_relative "../spec_helper"

# Characterization tests: pin CURRENT legacy behaviour (including quirks) to the
# golden fixtures. The Go service is validated against the same files.
FIXTURE_DIR = File.expand_path("../../../fixtures/golden/temp_tag", __dir__)

RSpec.describe "TempTagIssuer characterization" do
  Dir[File.join(FIXTURE_DIR, "*.json")].sort.each do |path|
    fx = JSON.parse(File.read(path))
    quirk = fx["tags"].any? { |t| t.start_with?("legacy_quirk") }

    it "#{fx["name"]}#{" [legacy quirk]" if quirk}" do
      actual = TempTagIssuer.new(Registration.new(fx["input"])).call
      expect(actual).to eq(fx["expected"])
    end
  end

  it "is deterministic for the same input" do
    fx = JSON.parse(File.read(File.join(FIXTURE_DIR, "CA_purchase_cash.json")))
    a = TempTagIssuer.new(Registration.new(fx["input"])).call
    b = TempTagIssuer.new(Registration.new(fx["input"])).call
    expect(a).to eq(b)
  end
end
