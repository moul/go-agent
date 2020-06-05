package filters

/*
- when the value is a hash:
  - matches if there is no valuePattern or keyPattern
    - expect(described_class.match_key_value({}, stage) { {} }).to eq(MATCH))
  - when an entry in the hash has a string value
    - matches if the key matches when there is only a keyPattern
      - expect(described_class.match_key_value(key_filter, stage) { { "key-match" => "test" } }).to eq(MATCH)
    - does NOT match if the key does NOT match when there is only a keyPattern
      - expect(described_class.match_key_value(key_filter, stage) { { "no match" => "value-match" } }).to eq(NO_MATCH)
    - matches if the value match when there is only a valuePattern
      - expect(described_class.match_key_value(value_filter, stage) { { "anything" => "value-match" } }).to eq(MATCH)
    - does NOT match if the value does NOT match when there is only a valuePattern
      - expect(described_class.match_key_value(value_filter, stage) { { "key-match" => "no match" } }).to eq(NO_MATCH)
    - matches if both the key and value match when there is both a keyPattern and valuePattern
      - expect(described_class.match_key_value(key_value_filter, stage) { { "key-match" => "value-match" } }).to eq(MATCH)
    - does NOT match if either the key or value do NOT match when there is both a keyPattern and valuePattern
      - expect(described_class.match_key_value(key_value_filter, stage) { { "key-match" => "no match" } }).to eq(NO_MATCH)
      - expect(described_class.match_key_value(key_value_filter, stage) { { "no match" => "value-match" } }).to eq(NO_MATCH)
  - when an entry in the hash is NOT a string value
    - matches if a hash value matches the filter (recursive)
      - expect(described_class.match_key_value(key_value_filter, stage) { {"anything" => {"key-match": "value-match"} } }).to eq(MATCH)
    - does NOT match if NO hash value matches
      - expect(described_class.match_key_value(value_filter, stage) { {"anything": 42} }).to eq(NO_MATCH)
      - expect(described_class.match_key_value(value_filter, stage) { {} }).to eq(NO_MATCH)

- when the value is an array
  - matches if there is no valuePattern or keyPattern
    - expect(described_class.match_key_value({}, stage) { [] }).to eq(MATCH)
  - matches if an element of the array matches the filter (recursive)
    - expect(described_class.match_key_value(value_filter, stage) { ["value-match"] }).to eq(MATCH)
  - does NOT match if NO element of the array matches
    - expect(described_class.match_key_value(value_filter, stage) { [42] }).to eq(NO_MATCH)
    - expect(described_class.match_key_value(value_filter, stage) { [] }).to eq(NO_MATCH)

- when the value is a string
  - matches if there is no valuePattern or keyPattern
    - expect(described_class.match_key_value({}, stage) { "anything" }).to eq(MATCH)
  - matches if there is a valuePattern and it matches the value
    - expect(described_class.match_key_value(value_filter, stage) { "value-match" }).to eq(MATCH)
  - does NOT match if there is valuePattern and it does NOT match the value
    - expect(described_class.match_key_value(value_filter, stage) { "no match" }).to eq(NO_MATCH)
  - does NOT match if there is a keyPattern
    - expect(described_class.match_key_value(key_value_filter, stage) { "value-match" }).to eq(NO_MATCH)

- when the value is a number
  - matches if there is no keyPattern or valuePattern
    - expect(described_class.match_key_value({}, stage) { 42 }).to eq(MATCH)
  - does NOT match otherwise
    - expect(described_class.match_key_value(value_filter, stage) { 42 }).to eq(NO_MATCH)

- when there is no value
  - does NOT match
    - expect(described_class.match_key_value({}, stage) { nil }).to eq(NO_MATCH)
 */
