package codec

import (
	"encoding/json/v2"
	"io"

	"github.com/fxamacker/cbor/v2"
)

var cborEncodeMode cbor.EncMode
var cborDecodeMode cbor.DecMode

func init() {
	var err error
	cborEncodeMode, err = (cbor.EncOptions{
		NilContainers: cbor.NilContainerAsEmpty,
		TextMarshaler: cbor.TextMarshalerTextString,
	}).EncMode()
	if err != nil {
		panic(err)
	}
	cborDecodeMode, err = (cbor.DecOptions{
		DupMapKey:                 cbor.DupMapKeyEnforcedAPF,
		TextUnmarshaler:           cbor.TextUnmarshalerTextString,
		JSONUnmarshalerTranscoder: _JSONScalarTranscoder{},
	}).DecMode()
	if err != nil {
		panic(err)
	}
}

// MarshalCBOR encodes a Skel payload, including native UUIDs as text strings.
func MarshalCBOR(value any) ([]byte, error) {
	return cborEncodeMode.Marshal(value)
}

// UnmarshalCBOR decodes a Skel payload into a non-nil pointer. JSON scalar
// decoders validate generated enums when no dedicated CBOR decoder exists.
// The target may be partially populated on error and must then be discarded.
func UnmarshalCBOR(data []byte, target any) error {
	return cborDecodeMode.Unmarshal(data, target)
}

type _JSONScalarTranscoder struct{}

func (_JSONScalarTranscoder) Transcode(w io.Writer, r io.Reader) error {
	var value any
	if err := cbor.NewDecoder(r).Decode(&value); err != nil {
		return err
	}
	return json.MarshalWrite(w, value)
}
