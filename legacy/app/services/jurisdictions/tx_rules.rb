module Jurisdictions
  class TxRules < BaseRules
    def code; "TX"; end
    def valid_days; 60; end
    def fee_cents; 500; end

    def required_documents(reg)
      d = %w[purchase_agreement form_130u]
      d << "lien_agreement" if reg.financed?
      d
    end

    def lien_errors(reg)
      return [] unless reg.financed?
      lh = reg.lienholder || {}
      (lh["name"] || lh[:name]).to_s.strip.empty? ? ["lienholder_required"] : []
    end

    def flags(reg)
      reg.attempt.to_i > 1 ? ["resubmission"] : []
    end

    # resubmissions: date from today's request
    def base_date(reg)
      reg.requested_on
    end
  end
end
