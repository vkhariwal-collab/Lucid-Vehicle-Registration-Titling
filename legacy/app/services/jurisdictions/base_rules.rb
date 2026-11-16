module Jurisdictions
  class BaseRules
    def code; "XX"; end
    def valid_days; 30; end
    def fee_cents; 0; end
    def currency; "USD"; end
    def required_documents(_reg); []; end
    def lien_errors(_reg); []; end
    def extra_errors(_reg); []; end
    def flags(_reg); ["unsupported_jurisdiction_default_rules"]; end
    def base_date(reg); reg.requested_on; end
    def adjust_expiry(d, _reg); [d, []]; end
  end

  def self.for(code)
    { "CA" => CaRules, "TX" => TxRules, "ON" => OnRules }.fetch(code.to_s, BaseRules).new
  end
end
