package codec_test

import (
	"testing"
	"uuid"

	"go.yorun.ai/skel/types/codec"
)

func TestJSONPayloadRules(t *testing.T) {
	type payload struct {
		ID     uuid.UUID         `json:"id"`
		Items  []string          `json:"items"`
		Lookup map[string]string `json:"lookup"`
	}
	input := payload{ID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")}
	data, err := codec.MarshalJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"id":"550e8400-e29b-41d4-a716-446655440000","items":[],"lookup":{}}` {
		t.Fatalf("unexpected wire payload: %s", data)
	}
	var result payload
	if err := codec.UnmarshalJSON(data, &result); err != nil {
		t.Fatal(err)
	}
	if result.ID != input.ID {
		t.Fatal("UUID changed")
	}
	if err := codec.UnmarshalJSON([]byte(`{"id":null,"id":null}`), &result); err == nil {
		t.Fatal("accepted duplicate keys")
	}
}

func TestMustMarshalJSONString(t *testing.T) {
	if got := codec.MustMarshalJSONString([]string(nil)); got != "[]" {
		t.Fatalf("nil collection must use shared JSON rules: %s", got)
	}
	t.Run("encoding failure panics", func(t *testing.T) {
		defer func() {
			if _, ok := recover().(error); !ok {
				t.Fatal("expected an encoding error panic")
			}
		}()
		codec.MustMarshalJSONString(make(chan int))
	})
}
