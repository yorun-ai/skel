package types

import (
	"encoding/base64"
	"encoding/json/v2"

	"github.com/fxamacker/cbor/v2"
)

// Binary encodes bytes as base64 in JSON and byte strings in CBOR.
type Binary []byte

func (b Binary) MarshalJSON() ([]byte, error) {
	return json.Marshal(base64.StdEncoding.EncodeToString([]byte(b)))
}

func (b *Binary) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*b = nil
		return nil
	}

	var encoded string
	if err := json.Unmarshal(data, &encoded); err != nil {
		return err
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return err
	}
	*b = Binary(decoded)
	return nil
}

func (b Binary) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal([]byte(b))
}

func (b *Binary) UnmarshalCBOR(data []byte) error {
	var decoded []byte
	if err := cbor.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*b = Binary(decoded)
	return nil
}
