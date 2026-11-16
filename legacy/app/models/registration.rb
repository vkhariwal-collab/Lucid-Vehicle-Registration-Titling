class Registration
  ATTRS = %i[vin jurisdiction delivery_jurisdiction purchase_type lienholder documents customer requested_at attempt original_issued_at state].freeze
  attr_accessor(*ATTRS)
  attr_reader :history

  def initialize(h = {})
    h = h.transform_keys(&:to_sym)
    ATTRS.each { |a| send("#{a}=", h[a]) }
    @documents ||= []
    @customer ||= {}
    @attempt ||= 1
    @state ||= "draft"
    @history = []
  end

  def financed?
    %w[finance lease].include?(purchase_type)
  end

  def lease?
    purchase_type == "lease"
  end

  def requested_on
    Date.parse(requested_at.to_s)
  end
end
