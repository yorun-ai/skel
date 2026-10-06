package types

import (
	"encoding/json/v2"
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"github.com/fxamacker/cbor/v2"
)

func TestTimestampJSONRoundTrip(t *testing.T) {
	type payload struct {
		Value Timestamp `json:"value"`
	}

	raw := time.Date(2026, 5, 4, 13, 14, 15, 123456789, time.FixedZone("CST+8", 8*3600))
	input := payload{
		Value: NewTimestamp(raw),
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if got := string(data); got != `{"value":"2026-05-04T05:14:15.123456789Z"}` {
		t.Fatalf("unexpected json: %s", got)
	}

	var decoded payload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if !decoded.Value.Equal(raw.UTC()) {
		t.Fatalf("unexpected timestamp value: got=%s want=%s", decoded.Value.Format(timestampLayout), raw.UTC().Format(timestampLayout))
	}
}

func TestTimestampCBORRoundTrip(t *testing.T) {
	type payload struct {
		Value Timestamp `cbor:"value"`
	}

	raw := time.Date(2026, 5, 4, 13, 14, 15, 123456789, time.FixedZone("CST+8", 8*3600))
	input := payload{
		Value: NewTimestamp(raw),
	}
	data, err := cbor.Marshal(input)
	if err != nil {
		t.Fatalf("MarshalCbor() error = %v", err)
	}
	got, err := cbor.Diagnose(data)
	if err != nil {
		t.Fatalf("Diagnose() error = %v", err)
	}
	if got != "{"+`"value": "2026-05-04T05:14:15.123456789Z"`+"}" {
		t.Fatalf("unexpected cbor diagnose: %s", got)
	}

	decoded, err := decodeCBOR[payload](data)
	if err != nil {
		t.Fatalf("UnmarshalCbor() error = %v", err)
	}
	if !decoded.Value.Equal(raw.UTC()) {
		t.Fatalf("unexpected timestamp value: got=%s want=%s", decoded.Value.Format(timestampLayout), raw.UTC().Format(timestampLayout))
	}
}

func TestNewTimestampNowReturnsUTCValue(t *testing.T) {
	value := NewTimestampNow()
	if value.Location() != time.UTC {
		t.Fatalf("unexpected timestamp location: got=%s want=%s", value.Location(), time.UTC)
	}
	if value.IsZero() {
		t.Fatal("unexpected zero timestamp")
	}
}

func TestLocalDateJSONRoundTrip(t *testing.T) {
	type payload struct {
		Value LocalDate `json:"value"`
	}

	input := payload{
		Value: NewLocalDate(civil.Date{
			Year:  2026,
			Month: time.May,
			Day:   4,
		}),
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if got := string(data); got != `{"value":"2026-05-04"}` {
		t.Fatalf("unexpected json: %s", got)
	}

	var decoded payload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if decoded.Value.Date != input.Value.Date {
		t.Fatalf("unexpected date value: got=%s want=%s", decoded.Value.String(), input.Value.String())
	}
}

func TestLocalDateCBORRoundTrip(t *testing.T) {
	type payload struct {
		Value LocalDate `cbor:"value"`
	}

	input := payload{
		Value: NewLocalDate(civil.Date{
			Year:  2026,
			Month: time.May,
			Day:   4,
		}),
	}
	data, err := cbor.Marshal(input)
	if err != nil {
		t.Fatalf("MarshalCbor() error = %v", err)
	}
	got, err := cbor.Diagnose(data)
	if err != nil {
		t.Fatalf("Diagnose() error = %v", err)
	}
	if got != "{"+`"value": "2026-05-04"`+"}" {
		t.Fatalf("unexpected cbor diagnose: %s", got)
	}

	decoded, err := decodeCBOR[payload](data)
	if err != nil {
		t.Fatalf("UnmarshalCbor() error = %v", err)
	}
	if decoded.Value.Date != input.Value.Date {
		t.Fatalf("unexpected date value: got=%s want=%s", decoded.Value.String(), input.Value.String())
	}
}

func TestLocalTimeJSONRoundTrip(t *testing.T) {
	type payload struct {
		Value LocalTime `json:"value"`
	}

	input := payload{
		Value: NewLocalTime(civil.Time{
			Hour:       9,
			Minute:     30,
			Second:     0,
			Nanosecond: 123000000,
		}),
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if got := string(data); got != `{"value":"09:30:00.123000000"}` {
		t.Fatalf("unexpected json: %s", got)
	}

	var decoded payload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if decoded.Value.Time != input.Value.Time {
		t.Fatalf("unexpected localtime value: got=%s want=%s", decoded.Value.String(), input.Value.String())
	}
}

func TestLocalDateTimeJSONRoundTrip(t *testing.T) {
	type payload struct {
		Value LocalDateTime `json:"value"`
	}

	input := payload{
		Value: NewLocalDateTime(civil.DateTime{
			Date: civil.Date{
				Year:  2026,
				Month: time.May,
				Day:   4,
			},
			Time: civil.Time{
				Hour:       9,
				Minute:     30,
				Second:     0,
				Nanosecond: 123000000,
			},
		}),
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if got := string(data); got != `{"value":"2026-05-04T09:30:00.123000000"}` {
		t.Fatalf("unexpected json: %s", got)
	}

	var decoded payload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if decoded.Value.DateTime != input.Value.DateTime {
		t.Fatalf("unexpected localdatetime value: got=%s want=%s", decoded.Value.String(), input.Value.String())
	}
}

func TestDurationJSONRoundTrip(t *testing.T) {
	type payload struct {
		Value Duration `json:"value"`
	}

	input := payload{
		Value: NewDuration(time.Hour + 2*time.Minute + 3*time.Second),
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if got := string(data); got != `{"value":"1h2m3s"}` {
		t.Fatalf("unexpected json: %s", got)
	}

	var decoded payload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if decoded.Value.Duration != input.Value.Duration {
		t.Fatalf("unexpected duration value: got=%s want=%s", decoded.Value.String(), input.Value.String())
	}
}

func TestDurationCBORRoundTrip(t *testing.T) {
	type payload struct {
		Value Duration `cbor:"value"`
	}

	input := payload{
		Value: NewDuration(250 * time.Millisecond),
	}
	data, err := cbor.Marshal(input)
	if err != nil {
		t.Fatalf("MarshalCbor() error = %v", err)
	}
	got, err := cbor.Diagnose(data)
	if err != nil {
		t.Fatalf("Diagnose() error = %v", err)
	}
	if got != "{"+`"value": "250ms"`+"}" {
		t.Fatalf("unexpected cbor diagnose: %s", got)
	}

	decoded, err := decodeCBOR[payload](data)
	if err != nil {
		t.Fatalf("UnmarshalCbor() error = %v", err)
	}
	if decoded.Value.Duration != input.Value.Duration {
		t.Fatalf("unexpected duration value: got=%s want=%s", decoded.Value.String(), input.Value.String())
	}
}
