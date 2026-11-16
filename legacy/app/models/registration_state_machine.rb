class RegistrationStateMachine
  class InvalidTransition < StandardError; end

  T = {
    submit_docs:    { from: %w[draft rejected], to: "docs_pending" },
    verify_docs:    { from: %w[docs_pending], to: "docs_complete" },
    issue_temp_tag: { from: %w[docs_complete], to: "temp_tag_issued" },
    submit_to_dmv:  { from: %w[temp_tag_issued], to: "submitted_to_dmv" },
    record_lien:    { from: %w[submitted_to_dmv], to: "lien_recorded" },
    complete_title: { from: %w[submitted_to_dmv lien_recorded], to: "titled" },
    reject:         { from: %w[docs_pending docs_complete temp_tag_issued submitted_to_dmv], to: "rejected" },
    escalate:       { from: %w[docs_pending docs_complete submitted_to_dmv], to: "manual_review" }
  }.freeze

  attr_reader :reg, :last_result

  def initialize(reg, elt: nil)
    @reg = reg
    @elt = elt
  end

  def fire!(ev)
    t = T.fetch(ev.to_sym) { raise InvalidTransition, "unknown event #{ev}" }
    raise InvalidTransition, "#{ev} not allowed from #{reg.state}" unless t[:from].include?(reg.state)
    send("g_#{ev}") if respond_to?("g_#{ev}", true)
    prev = reg.state
    reg.state = t[:to]
    reg.history << [prev, ev.to_s, reg.state]
    send("a_#{ev}") if respond_to?("a_#{ev}", true)
    reg.state
  end

  private

  def g_issue_temp_tag
    @last_result = TempTagIssuer.new(reg).call
    raise InvalidTransition, "temp tag rejected: #{@last_result["reasons"].join(",")}" if @last_result["status"] != "issued"
  end

  def g_record_lien
    raise InvalidTransition, "record_lien requires financed purchase" unless reg.financed?
    raise InvalidTransition, "record_lien not used for ON" if reg.jurisdiction == "ON"
  end

  def g_complete_title
    if reg.financed? && reg.jurisdiction != "ON" && reg.state != "lien_recorded"
      raise InvalidTransition, "lien must be recorded before title"
    end
  end

  def a_record_lien
    @elt&.file!(reg)
  end
end
