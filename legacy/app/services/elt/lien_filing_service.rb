module Elt
  class TimeoutError < StandardError; end

  class LienFilingService
    attr_reader :calls

    def initialize(client)
      @client = client
      @calls = 0
    end

    def file!(reg)
      tries = 0
      begin
        tries += 1
        @calls += 1
        @client.post("/liens", vin: reg.vin, lienholder: reg.lienholder, jurisdiction: reg.jurisdiction)
      rescue TimeoutError
        retry if tries < 3
        raise
      end
    end
  end
end
