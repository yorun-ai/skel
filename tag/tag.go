// Package tag reads Skel metadata from Go struct tags. It is independent of
// compiler models and application runtimes; callers apply the parsed metadata.
package tag

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// IsSensitive reports whether the skel tag marks the field as sensitive.
// Attributes match exactly, ignoring surrounding whitespace.
func IsSensitive(value reflect.StructTag) bool {
	return hasFlag(value, "sensitive")
}

// IsIdentifier reports whether the skel tag marks the field as an actor identifier.
// Attributes match exactly, ignoring surrounding whitespace.
func IsIdentifier(value reflect.StructTag) bool {
	return hasFlag(value, "identifier")
}

// Index reads index(n) from the skel tag, ignoring unrelated attributes.
// Malformed or repeated indexes are errors. It does not read legacy arg tags;
// callers validate the index against the argument structure's size.
func Index(value reflect.StructTag) (index int, found bool, err error) {
	for option := range strings.SplitSeq(value.Get("skel"), ",") {
		option = strings.TrimSpace(option)
		if option != "index" && !strings.HasPrefix(option, "index(") {
			continue
		}
		if found {
			return 0, false, fmt.Errorf("duplicate skel index attribute")
		}
		argument, ok := strings.CutPrefix(option, "index(")
		if !ok || !strings.HasSuffix(argument, ")") {
			return 0, false, fmt.Errorf("invalid skel index attribute %q", option)
		}
		index, err = strconv.Atoi(strings.TrimSuffix(argument, ")"))
		if err != nil {
			return 0, false, fmt.Errorf("invalid skel index attribute %q: %w", option, err)
		}
		found = true
	}
	return index, found, nil
}

func hasFlag(value reflect.StructTag, name string) bool {
	for option := range strings.SplitSeq(value.Get("skel"), ",") {
		if strings.TrimSpace(option) == name {
			return true
		}
	}
	return false
}
