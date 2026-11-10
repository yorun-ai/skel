package codec_test

import (
	"bytes"
	"encoding/json/v2"
	"fmt"
	"reflect"
	"testing"
	"time"
	"uuid"

	"github.com/fxamacker/cbor/v2"
	"go.yorun.ai/skel/types"
	"go.yorun.ai/skel/types/codec"
)

func TestCBORNativeUUIDCompatibility(t *testing.T) {
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	//lint:ignore SA1019 Verify wire compatibility with the deprecated UUID representation.
	old := map[types.UUID][]types.UUID{types.NewUUID(id): {types.NewUUID(id), {}}}
	native := map[uuid.UUID][]uuid.UUID{id: {id, {}}}
	before, err := cbor.Marshal(old)
	if err != nil {
		t.Fatal(err)
	}
	after, err := codec.MarshalCBOR(native)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatalf("UUID wire representation changed: %x != %x", before, after)
	}
	var result map[uuid.UUID][]uuid.UUID
	if err := codec.UnmarshalCBOR(before, &result); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(native, result) {
		t.Fatalf("unexpected decoded value: %#v", result)
	}
}

func TestCBORRetainsScalarEncodings(t *testing.T) {
	for _, value := range []any{types.Decimal{}, types.Timestamp{}, types.Duration{}, types.LocalTime{}, types.NewLocalDateOf(time.Date(2026, 11, 11, 0, 0, 0, 0, time.UTC)), types.NewLocalDateTimeOf(time.Date(2026, 11, 11, 12, 34, 56, 0, time.UTC)), types.Binary{0, 255}, types.JSON(`{"a":1}`)} {
		t.Run(fmt.Sprintf("%T", value), func(t *testing.T) {
			before, err := cbor.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			after, err := codec.MarshalCBOR(value)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, after) {
				t.Fatalf("encoding changed: %x != %x", before, after)
			}
			oldResult := reflect.New(reflect.TypeOf(value)).Interface()
			newResult := reflect.New(reflect.TypeOf(value)).Interface()
			if err := cbor.Unmarshal(before, oldResult); err != nil {
				t.Fatal(err)
			}
			if err := codec.UnmarshalCBOR(before, newResult); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(oldResult, newResult) {
				t.Fatalf("decoding changed: %#v != %#v", oldResult, newResult)
			}
		})
	}
}

type validatedEnum string

func (e *validatedEnum) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return fmt.Errorf("enum cannot be null")
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if value != "ACTIVE" {
		value = "UNSPECIFIED"
	}
	*e = validatedEnum(value)
	return nil
}

func TestCBORUsesJSONScalarValidation(t *testing.T) {
	for _, raw := range []any{nil, "ACTIVE", "UNKNOWN"} {
		data, err := cbor.Marshal(raw)
		if err != nil {
			t.Fatal(err)
		}
		var result validatedEnum
		err = codec.UnmarshalCBOR(data, &result)
		if raw == nil {
			if err == nil {
				t.Fatal("accepted null enum")
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		want := validatedEnum("UNSPECIFIED")
		if raw == "ACTIVE" {
			want = "ACTIVE"
		}
		if result != want {
			t.Fatalf("got %s, want %s", result, want)
		}
	}
}

func TestCBORCollectionRules(t *testing.T) {
	for _, value := range []any{[]string(nil), map[string]string(nil)} {
		data, err := codec.MarshalCBOR(value)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(data, []byte{0x80}) && !bytes.Equal(data, []byte{0xa0}) {
			t.Fatalf("nil collection encoded as %x", data)
		}
	}
	var target map[string]int
	if err := codec.UnmarshalCBOR([]byte{0xa2, 0x61, 'a', 1, 0x61, 'a', 2}, &target); err == nil {
		t.Fatal("accepted duplicate keys")
	}
}
