package types

import (
	"encoding/json/v2"
	"time"

	"cloud.google.com/go/civil"
	"github.com/fxamacker/cbor/v2"
)

const timestampLayout = time.RFC3339Nano

// Timestamp is the shared skel timestamp type.
// It is encoded as an RFC3339Nano string in both JSON and CBOR.
type Timestamp struct {
	time.Time
}

// NewTimestamp converts t to the Skel timestamp representation.
func NewTimestamp(t time.Time) Timestamp {
	return Timestamp{
		Time: t.UTC(),
	}
}

// NewTimestampNow returns the current UTC timestamp.
func NewTimestampNow() Timestamp {
	return NewTimestamp(time.Now())
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.UTC().Format(timestampLayout))
}

func (t *Timestamp) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		t.Time = time.Time{}
		return nil
	}

	var encoded string
	if err := json.Unmarshal(data, &encoded); err != nil {
		return err
	}

	decoded, err := time.Parse(timestampLayout, encoded)
	if err != nil {
		return err
	}
	t.Time = decoded.UTC()
	return nil
}

func (t Timestamp) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(t.UTC().Format(timestampLayout))
}

func (t *Timestamp) UnmarshalCBOR(data []byte) error {
	var encoded *string
	if err := cbor.Unmarshal(data, &encoded); err != nil {
		return err
	}
	if encoded == nil {
		t.Time = time.Time{}
		return nil
	}

	decoded, err := time.Parse(timestampLayout, *encoded)
	if err != nil {
		return err
	}
	t.Time = decoded.UTC()
	return nil
}

// Duration is the shared skel duration type.
// It is encoded as a time.ParseDuration-compatible string in both JSON and CBOR.
type Duration struct {
	time.Duration
}

// NewDuration converts d to the Skel duration representation.
func NewDuration(d time.Duration) Duration {
	return Duration{
		Duration: d,
	}
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		d.Duration = 0
		return nil
	}

	var encoded string
	if err := json.Unmarshal(data, &encoded); err != nil {
		return err
	}

	decoded, err := time.ParseDuration(encoded)
	if err != nil {
		return err
	}
	d.Duration = decoded
	return nil
}

func (d Duration) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(d.String())
}

func (d *Duration) UnmarshalCBOR(data []byte) error {
	var encoded *string
	if err := cbor.Unmarshal(data, &encoded); err != nil {
		return err
	}
	if encoded == nil {
		d.Duration = 0
		return nil
	}

	decoded, err := time.ParseDuration(*encoded)
	if err != nil {
		return err
	}
	d.Duration = decoded
	return nil
}

// LocalDate is the shared skel local date type.
// It is encoded as an RFC3339 full-date string in both JSON and CBOR.
type LocalDate struct {
	civil.Date
}

// NewLocalDate converts date to the Skel local-date representation.
func NewLocalDate(date civil.Date) LocalDate {
	return LocalDate{
		Date: date,
	}
}

// NewLocalDateOf extracts the local date from t.
func NewLocalDateOf(t time.Time) LocalDate {
	return NewLocalDate(civil.DateOf(t))
}

func (d LocalDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *LocalDate) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		d.Date = civil.Date{}
		return nil
	}

	var encoded string
	if err := json.Unmarshal(data, &encoded); err != nil {
		return err
	}

	decoded, err := civil.ParseDate(encoded)
	if err != nil {
		return err
	}
	d.Date = decoded
	return nil
}

func (d LocalDate) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(d.String())
}

func (d *LocalDate) UnmarshalCBOR(data []byte) error {
	var encoded *string
	if err := cbor.Unmarshal(data, &encoded); err != nil {
		return err
	}
	if encoded == nil {
		d.Date = civil.Date{}
		return nil
	}

	decoded, err := civil.ParseDate(*encoded)
	if err != nil {
		return err
	}
	d.Date = decoded
	return nil
}

// LocalTime is the shared skel local time type.
// It is encoded as an RFC3339 partial-time string in both JSON and CBOR.
type LocalTime struct {
	civil.Time
}

// NewLocalTime converts clock to the Skel local-time representation.
func NewLocalTime(clock civil.Time) LocalTime {
	return LocalTime{
		Time: clock,
	}
}

// NewLocalTimeOf extracts the local time from t.
func NewLocalTimeOf(t time.Time) LocalTime {
	return NewLocalTime(civil.TimeOf(t))
}

func (t LocalTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

func (t *LocalTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		t.Time = civil.Time{}
		return nil
	}

	var encoded string
	if err := json.Unmarshal(data, &encoded); err != nil {
		return err
	}

	decoded, err := civil.ParseTime(encoded)
	if err != nil {
		return err
	}
	t.Time = decoded
	return nil
}

func (t LocalTime) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(t.String())
}

func (t *LocalTime) UnmarshalCBOR(data []byte) error {
	var encoded *string
	if err := cbor.Unmarshal(data, &encoded); err != nil {
		return err
	}
	if encoded == nil {
		t.Time = civil.Time{}
		return nil
	}

	decoded, err := civil.ParseTime(*encoded)
	if err != nil {
		return err
	}
	t.Time = decoded
	return nil
}

// LocalDateTime is the shared skel local datetime type.
// It is encoded as an RFC3339 date-time string without timezone in both JSON and CBOR.
type LocalDateTime struct {
	civil.DateTime
}

// NewLocalDateTime converts dateTime to the Skel local-date-time representation.
func NewLocalDateTime(dateTime civil.DateTime) LocalDateTime {
	return LocalDateTime{
		DateTime: dateTime,
	}
}

// NewLocalDateTimeOf extracts the local date and time from t.
func NewLocalDateTimeOf(t time.Time) LocalDateTime {
	return NewLocalDateTime(civil.DateTimeOf(t))
}

func (dt LocalDateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(dt.String())
}

func (dt *LocalDateTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		dt.DateTime = civil.DateTime{}
		return nil
	}

	var encoded string
	if err := json.Unmarshal(data, &encoded); err != nil {
		return err
	}

	decoded, err := civil.ParseDateTime(encoded)
	if err != nil {
		return err
	}
	dt.DateTime = decoded
	return nil
}

func (dt LocalDateTime) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(dt.String())
}

func (dt *LocalDateTime) UnmarshalCBOR(data []byte) error {
	var encoded *string
	if err := cbor.Unmarshal(data, &encoded); err != nil {
		return err
	}
	if encoded == nil {
		dt.DateTime = civil.DateTime{}
		return nil
	}

	decoded, err := civil.ParseDateTime(*encoded)
	if err != nil {
		return err
	}
	dt.DateTime = decoded
	return nil
}
