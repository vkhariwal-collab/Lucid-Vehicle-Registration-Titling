module Jurisdictions
  class CaRules < BaseRules
    def code; "CA"; end
    def valid_days; 90; end

    def required_documents(reg)
      d = %w[purchase_agreement odometer_disclosure]
      d << "lien_agreement" if reg.financed?
      d
    end

    def lien_errors(reg)
      return [] unless reg.financed?
      lh = reg.lienholder || {}
      (lh["elt_id"] || lh[:elt_id]).to_s.empty? ? ["elt_lienholder_required"] : []
    end

    def flags(_reg); []; end

    def adjust_expiry(d, _reg)
      d.sunday? ? [d + 1, ["expiry_rolled_from_sunday"]] : [d, []]
    end
  end
end
