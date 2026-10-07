package types

import (
	"encoding/json/v2"
	"testing"

	"github.com/fxamacker/cbor/v2"
)

func TestBinaryJSONRoundTrip(t *testing.T) {
	type payload struct {
		Value Binary `json:"value"`
	}

	input := payload{
		Value: Binary([]byte("hello")),
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if got := string(data); got != `{"value":"aGVsbG8="}` {
		t.Fatalf("unexpected json: %s", got)
	}

	var decoded payload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if string(decoded.Value) != "hello" {
		t.Fatalf("unexpected binary value: %q", string(decoded.Value))
	}
}

func TestBinaryUnmarshalJSONNull(t *testing.T) {
	var value Binary
	if err := json.Unmarshal([]byte("null"), &value); err != nil {
		t.Fatalf("Unmarshal(null) error = %v", err)
	}
	if value != nil {
		t.Fatalf("expected nil binary, got %#v", []byte(value))
	}
}

func TestBinaryCBORRoundTrip(t *testing.T) {
	type payload struct {
		Value Binary `cbor:"value"`
	}

	input := payload{
		Value: Binary([]byte("hello")),
	}
	data, err := cbor.Marshal(input)
	if err != nil {
		t.Fatalf("MarshalCbor() error = %v", err)
	}

	decoded, err := decodeCBOR[payload](data)
	if err != nil {
		t.Fatalf("UnmarshalCbor() error = %v", err)
	}
	if string(decoded.Value) != "hello" {
		t.Fatalf("unexpected binary value: %q", string(decoded.Value))
	}
}
