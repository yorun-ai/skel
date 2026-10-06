// Package types provides Go representations of Skel scalar values and their
// JSON and CBOR encodings, independently of the compiler and application runtimes.
//
// Decimal, timestamps, durations, local dates and times, UUIDs, and JSON text
// encode as strings in both formats. Binary encodes as base64 in JSON and as
// bytes in CBOR. Native Go types represent the remaining Skel scalars.
// Keeping the JSON and CBOR payload shapes aligned preserves shared protocol
// semantics across language bindings; Binary uses each format's native byte
// representation where available.
package types
