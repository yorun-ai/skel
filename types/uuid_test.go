package types

import (
	"encoding/json/v2"
	"testing"
	"uuid"

	"github.com/fxamacker/cbor/v2"
)

func TestUUIDJSONRoundTrip(t *testing.T) {
	type payload struct {
		Value UUID `json:"value"`
	}

	input := payload{
		Value: NewUUID(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")),
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if got := string(data); got != `{"value":"550e8400-e29b-41d4-a716-446655440000"}` {
		t.Fatalf("unexpected json: %s", got)
	}

	var decoded payload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if decoded.Value.UUID != input.Value.UUID {
		t.Fatalf("unexpected uuid value: got=%s want=%s", decoded.Value.String(), input.Value.String())
	}
}

func TestUUIDCBORRoundTrip(t *testing.T) {
	type payload struct {
		Value UUID `cbor:"value"`
	}

	input := payload{
		Value: NewUUID(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")),
	}
	data, err := cbor.Marshal(input)
	if err != nil {
		t.Fatalf("MarshalCbor() error = %v", err)
	}

	decoded, err := decodeCBOR[payload](data)
	if err != nil {
		t.Fatalf("UnmarshalCbor() error = %v", err)
	}
	if decoded.Value.UUID != input.Value.UUID {
		t.Fatalf("unexpected uuid value: got=%s want=%s", decoded.Value.String(), input.Value.String())
	}
}

func TestUUIDMapKeyJSONRoundTrip(t *testing.T) {
	type payload struct {
		Values map[UUID]string `json:"values"`
	}

	id := NewUUID(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"))
	input := payload{
		Values: map[UUID]string{id: "value"},
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if got := string(data); got != `{"values":{"550e8400-e29b-41d4-a716-446655440000":"value"}}` {
		t.Fatalf("unexpected json: %s", got)
	}

	var decoded payload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got := decoded.Values[id]; got != "value" {
		t.Fatalf("unexpected map value: %q", got)
	}
}

func TestUUIDMapKeyCBORRoundTrip(t *testing.T) {
	type payload struct {
		Values map[UUID]string `cbor:"values"`
	}

	id := NewUUID(uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"))
	input := payload{
		Values: map[UUID]string{id: "value"},
	}
	data, err := cbor.Marshal(input)
	if err != nil {
		t.Fatalf("MarshalCbor() error = %v", err)
	}

	decoded, err := decodeCBOR[payload](data)
	if err != nil {
		t.Fatalf("UnmarshalCbor() error = %v", err)
	}
	if got := decoded.Values[id]; got != "value" {
		t.Fatalf("unexpected map value: %q", got)
	}
}
