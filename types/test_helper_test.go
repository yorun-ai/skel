package types

import (
	"github.com/fxamacker/cbor/v2"
)

func decodeCBOR[T any](data []byte) (T, error) {
	var result T
	err := cbor.Unmarshal(data, &result)
	return result, err
}
