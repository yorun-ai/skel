package types

import (
	"encoding/json/v2"

	"github.com/fxamacker/cbor/v2"
	"github.com/shopspring/decimal"
)

// Decimal is the shared skel decimal type.
// It is encoded as a decimal string in both JSON and CBOR.
type Decimal struct {
	decimal.Decimal
}

// NewDecimal converts value to the Skel decimal representation.
func NewDecimal(value decimal.Decimal) Decimal {
	return Decimal{
		Decimal: value,
	}
}

// encodeString keeps the decimal scale in the wire representation. decimal.String()
// normalizes trailing zeros, so values like "1.00" would otherwise become "1".
func (d Decimal) encodeString() string {
	exponent := d.Exponent()
	if exponent < 0 {
		return d.StringFixed(-exponent)
	}
	return d.String()
}

func (d Decimal) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.encodeString())
}

func (d *Decimal) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		d.Decimal = decimal.Decimal{}
		return nil
	}

	var encoded string
	if err := json.Unmarshal(data, &encoded); err != nil {
		return err
	}

	decoded, err := decimal.NewFromString(encoded)
	if err != nil {
		return err
	}
	d.Decimal = decoded
	return nil
}

func (d Decimal) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(d.encodeString())
}

func (d *Decimal) UnmarshalCBOR(data []byte) error {
	var encoded *string
	if err := cbor.Unmarshal(data, &encoded); err != nil {
		return err
	}
	if encoded == nil {
		d.Decimal = decimal.Decimal{}
		return nil
	}

	decoded, err := decimal.NewFromString(*encoded)
	if err != nil {
		return err
	}
	d.Decimal = decoded
	return nil
}
