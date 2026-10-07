package types

import (
	"encoding/json/v2"

	"github.com/fxamacker/cbor/v2"
)

// JSON is the shared skel JSON text type.
// It is encoded as a JSON string in both JSON and CBOR.
type JSON string

func (j JSON) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(j))
}

func (j *JSON) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*j = ""
		return nil
	}
	var encoded string
	if err := json.Unmarshal(data, &encoded); err != nil {
		return err
	}
	*j = JSON(encoded)
	return nil
}

func (j JSON) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(string(j))
}

func (j *JSON) UnmarshalCBOR(data []byte) error {
	var encoded *string
	if err := cbor.Unmarshal(data, &encoded); err != nil {
		return err
	}
	if encoded == nil {
		*j = ""
		return nil
	}
	*j = JSON(*encoded)
	return nil
}
