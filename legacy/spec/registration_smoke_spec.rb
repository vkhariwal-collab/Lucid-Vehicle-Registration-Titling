require_relative "spec_helper"

RSpec.describe TempTagIssuer do
  it "issues a CA tag" do
    reg = Registration.new(vin: "5L1RA1AA0SA000001", jurisdiction: "CA", purchase_type: "cash",
                           documents: %w[purchase_agreement odometer_disclosure], requested_at: "2026-11-02")
    expect(TempTagIssuer.new(reg).call["status"]).to eq("issued")
  end
end
