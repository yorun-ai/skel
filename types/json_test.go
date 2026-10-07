package types

import (
	"encoding/json/v2"
	"testing"

	"github.com/fxamacker/cbor/v2"
)

func TestJSONJSONRoundTrip(t *testing.T) {
	type payload struct {
		Value JSON `json:"value"`
	}

	input := payload{
		Value: JSON(`{"name":"vine","count":2}`),
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if got := string(data); got != `{"value":"{\"name\":\"vine\",\"count\":2}"}` {
		t.Fatalf("unexpected json: %s", got)
	}

	var decoded payload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if decoded.Value != input.Value {
		t.Fatalf("unexpected json value: got=%s want=%s", decoded.Value, input.Value)
	}
}

func TestJSONCBORRoundTrip(t *testing.T) {
	type payload struct {
		Value JSON `cbor:"value"`
	}

	input := payload{
		Value: JSON(`{"name":"vine","count":2}`),
	}
	data, err := cbor.Marshal(input)
	if err != nil {
		t.Fatalf("MarshalCbor() error = %v", err)
	}

	decoded, err := decodeCBOR[payload](data)
	if err != nil {
		t.Fatalf("UnmarshalCbor() error = %v", err)
	}
	if decoded.Value != input.Value {
		t.Fatalf("unexpected json value: got=%s want=%s", decoded.Value, input.Value)
	}
}
