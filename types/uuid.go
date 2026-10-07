package types

import (
	"encoding/json/v2"
	"uuid"

	"github.com/fxamacker/cbor/v2"
)

// UUID is the shared skel UUID type.
// It is encoded as a UUID string in both JSON and CBOR.
type UUID struct {
	uuid.UUID
}

// NewUUID converts id to the Skel UUID representation.
func NewUUID(id uuid.UUID) UUID {
	return UUID{
		UUID: id,
	}
}

func (u UUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.String())
}

func (u *UUID) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		u.UUID = uuid.UUID{}
		return nil
	}

	var encoded string
	if err := json.Unmarshal(data, &encoded); err != nil {
		return err
	}

	decoded, err := uuid.Parse(encoded)
	if err != nil {
		return err
	}
	u.UUID = decoded
	return nil
}

func (u UUID) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(u.String())
}

func (u *UUID) UnmarshalCBOR(data []byte) error {
	var encoded *string
	if err := cbor.Unmarshal(data, &encoded); err != nil {
		return err
	}
	if encoded == nil {
		u.UUID = uuid.UUID{}
		return nil
	}

	decoded, err := uuid.Parse(*encoded)
	if err != nil {
		return err
	}
	u.UUID = decoded
	return nil
}
