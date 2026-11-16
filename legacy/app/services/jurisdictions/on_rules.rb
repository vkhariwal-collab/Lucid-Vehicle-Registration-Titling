module Jurisdictions
  class OnRules < BaseRules
    POSTAL = /\A[A-Z]\d[A-Z] ?\d[A-Z]\d\z/

    def code; "ON"; end
    def valid_days; 10; end
    def fee_cents; 3200; end
    def currency; "CAD"; end
    def required_documents(_reg); %w[bill_of_sale proof_of_insurance]; end

    def extra_errors(reg)
      pc = reg.customer["postal_code"] || reg.customer[:postal_code]
      pc.to_s =~ POSTAL ? [] : ["invalid_postal_code"]
    end

    def flags(reg)
      f = []
      f << "ppsa_registration_required" if reg.financed?
      f << "lessor_owned" if reg.lease?
      f
    end
  end
end
