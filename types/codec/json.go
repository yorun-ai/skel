// Package codec defines the shared JSON and CBOR payload encoding rules for
// Skel values. It does not encode transport envelopes or apply body limits.
// Native UUIDs encode as strings, nil collections encode as empty collections,
// and decoding rejects duplicate keys. Types with their own encoding methods
// retain those methods; CBOR decoding also honors JSON scalar validation.
package codec

import "encoding/json/v2"

// MarshalJSON encodes a Skel payload with the shared JSON rules.
func MarshalJSON(value any) ([]byte, error) {
	return json.Marshal(value)
}

// MustMarshalJSONString encodes a Skel payload as a JSON string and panics on failure.
func MustMarshalJSONString(value any) string {
	data, err := MarshalJSON(value)
	if err != nil {
		panic(err)
	}
	return string(data)
}

// UnmarshalJSON decodes a Skel payload into a non-nil pointer.
// The target may be partially populated on error and must then be discarded.
func UnmarshalJSON(data []byte, target any) error {
	return json.Unmarshal(data, target)
}
