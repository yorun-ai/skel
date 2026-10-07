package types

import (
	"encoding/json/v2"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/shopspring/decimal"
)

func TestDecimalJSONRoundTrip(t *testing.T) {
	type payload struct {
		Value Decimal `json:"value"`
	}

	input := payload{
		Value: NewDecimal(decimal.RequireFromString("1.00")),
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if got := string(data); got != `{"value":"1.00"}` {
		t.Fatalf("unexpected json: %s", got)
	}

	var decoded payload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if !decoded.Value.Equal(input.Value.Decimal) {
		t.Fatalf("unexpected decimal value: got=%s want=%s", decoded.Value.String(), input.Value.String())
	}
}

func TestDecimalCBORRoundTrip(t *testing.T) {
	type payload struct {
		Value Decimal `cbor:"value"`
	}

	input := payload{
		Value: NewDecimal(decimal.RequireFromString("1.00")),
	}
	data, err := cbor.Marshal(input)
	if err != nil {
		t.Fatalf("MarshalCbor() error = %v", err)
	}
	var encoded map[string]string
	if err := cbor.Unmarshal(data, &encoded); err != nil {
		t.Fatalf("Unmarshal() encoded cbor error = %v", err)
	}
	if got := encoded["value"]; got != "1.00" {
		t.Fatalf("unexpected cbor decimal text: %s", got)
	}

	decoded, err := decodeCBOR[payload](data)
	if err != nil {
		t.Fatalf("UnmarshalCbor() error = %v", err)
	}
	if !decoded.Value.Equal(input.Value.Decimal) {
		t.Fatalf("unexpected decimal value: got=%s want=%s", decoded.Value.String(), input.Value.String())
	}
}
