class TempTagIssuer
  VIN_RE = /\A[A-HJ-NPR-Z0-9]{17}\z/

  def initialize(reg)
    @reg = reg
  end

  def call
    return rej(["invalid_vin"]) unless @reg.vin.to_s =~ VIN_RE
    r = Jurisdictions.for(@reg.jurisdiction)
    errs = (r.required_documents(@reg) - @reg.documents.map(&:to_s)).map { |d| "missing_document:#{d}" }
    errs += r.lien_errors(@reg)
    errs += r.extra_errors(@reg)
    return rej(errs) if errs.any?

    exp, f2 = r.adjust_expiry(r.base_date(@reg) + r.valid_days, @reg)
    {
      "status" => "issued",
      "tag_number" => tag(r),
      "expires_on" => exp.iso8601,
      "valid_days" => r.valid_days,
      "fee_cents" => r.fee_cents,
      "currency" => r.currency,
      "reasons" => [],
      "flags" => r.flags(@reg) + f2
    }
  end

  private

  def tag(r)
    "#{r.code}-#{Digest::SHA256.hexdigest("#{@reg.vin}:#{@reg.jurisdiction}:#{@reg.attempt}")[0, 8].upcase}"
  end

  def rej(reasons)
    { "status" => "rejected", "tag_number" => nil, "expires_on" => nil, "valid_days" => 0,
      "fee_cents" => 0, "currency" => nil, "reasons" => reasons, "flags" => [] }
  end
end
