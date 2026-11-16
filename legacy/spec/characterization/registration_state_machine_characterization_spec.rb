require_relative "../spec_helper"

RSpec.describe "RegistrationStateMachine characterization" do
  let(:ca_cash) do
    { vin: "5L1RA1AA0SA000101", jurisdiction: "CA", purchase_type: "cash",
      documents: %w[purchase_agreement odometer_disclosure], requested_at: "2026-11-03" }
  end
  let(:ca_fin) do
    ca_cash.merge(vin: "5L1RA1AA0SA000102", purchase_type: "finance",
                  lienholder: { "name" => "LFS", "elt_id" => "ELT-CA-0042" },
                  documents: %w[purchase_agreement odometer_disclosure lien_agreement])
  end
  let(:on_fin) do
    { vin: "5L1RA1AA0SA000302", jurisdiction: "ON", purchase_type: "finance",
      lienholder: { "name" => "LFS Canada" }, documents: %w[bill_of_sale proof_of_insurance],
      customer: { "postal_code" => "K1A0B1" }, requested_at: "2026-11-06" }
  end

  def run(attrs, events, elt: nil)
    reg = Registration.new(attrs)
    sm = RegistrationStateMachine.new(reg, elt: elt)
    events.each { |e| sm.fire!(e) }
    reg
  end

  HAPPY = %i[submit_docs verify_docs issue_temp_tag submit_to_dmv].freeze

  it "cash purchase: titles straight from submitted_to_dmv" do
    reg = run(ca_cash, HAPPY + [:complete_title])
    expect(reg.state).to eq("titled")
    expect(reg.history.map(&:last)).to eq(%w[docs_pending docs_complete temp_tag_issued submitted_to_dmv titled])
  end

  it "CA financed: title is blocked until lien recorded" do
    reg = run(ca_fin, HAPPY)
    sm = RegistrationStateMachine.new(reg)
    expect { sm.fire!(:complete_title) }.to raise_error(RegistrationStateMachine::InvalidTransition, /lien must be recorded/)
    sm.fire!(:record_lien)
    expect(sm.fire!(:complete_title)).to eq("titled")
  end

  it "ON financed: skips lien_recorded entirely (PPSA handled elsewhere) [legacy quirk]" do
    reg = run(on_fin, HAPPY + [:complete_title])
    expect(reg.state).to eq("titled")
    sm = RegistrationStateMachine.new(Registration.new(on_fin.merge(state: "submitted_to_dmv")))
    expect { sm.fire!(:record_lien) }.to raise_error(RegistrationStateMachine::InvalidTransition, /not used for ON/)
  end

  it "issue_temp_tag guard leaves state unchanged when the issuer rejects" do
    reg = Registration.new(ca_cash.merge(documents: ["purchase_agreement"], state: "docs_complete"))
    sm = RegistrationStateMachine.new(reg)
    expect { sm.fire!(:issue_temp_tag) }.to raise_error(/missing_document:odometer_disclosure/)
    expect(reg.state).to eq("docs_complete")
  end

  it "rejected registrations can be resubmitted" do
    reg = run(ca_cash, %i[submit_docs reject submit_docs])
    expect(reg.state).to eq("docs_pending")
  end

  {
    "draft" => %i[verify_docs issue_temp_tag submit_to_dmv record_lien complete_title reject escalate],
    "titled" => %i[submit_docs reject escalate record_lien]
  }.each do |state, events|
    events.each do |ev|
      it "rejects #{ev} from #{state}" do
        sm = RegistrationStateMachine.new(Registration.new(ca_fin.merge(state: state)))
        expect { sm.fire!(ev) }.to raise_error(RegistrationStateMachine::InvalidTransition)
      end
    end
  end

  describe "ELT lien filing" do
    let(:flaky) do
      Class.new do
        attr_reader :posts
        def initialize; @posts = []; @n = 0; end
        def post(path, body)
          @posts << [path, body]
          @n += 1
          raise Elt::TimeoutError if @n < 3
          { "status" => "accepted" }
        end
      end.new
    end

    it "retries timeouts 3x WITHOUT an idempotency key -> provider sees 3 lien posts [legacy quirk: duplicate-lien risk]" do
      elt = Elt::LienFilingService.new(flaky)
      reg = Registration.new(ca_fin.merge(state: "submitted_to_dmv"))
      RegistrationStateMachine.new(reg, elt: elt).fire!(:record_lien)
      expect(flaky.posts.size).to eq(3)
      expect(flaky.posts.map(&:last).uniq.size).to eq(1)
      expect(flaky.posts.first.last.keys).not_to include(:idempotency_key)
    end
  end
end
