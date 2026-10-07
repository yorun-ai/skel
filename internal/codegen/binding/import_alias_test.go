package binding

import (
	"reflect"
	"strings"
	"testing"
)

func TestResolveImportAliases(t *testing.T) {
	for _, reversed := range []bool{false, true} {
		types := []*ImportBinding{
			{Domain: "first.user", Alias: "userapi"},
			{Domain: "second.user", Alias: "userapi"},
			{Domain: "explicit", Alias: "firstuser", Explicit: true},
			{Domain: "plain.order", Alias: "orderapi"},
			{Domain: "seconduser", Alias: "context", Explicit: true},
		}
		if reversed {
			for i, j := 0, len(types)-1; i < j; i, j = i+1, j-1 {
				types[i], types[j] = types[j], types[i]
			}
		}
		ResolveImportAliases(types, func(domain string) string { return strings.ReplaceAll(domain, ".", "") }, []string{"context"})
		got := map[string]string{}
		for _, kind := range types {
			got[kind.Domain] = kind.Alias
		}
		want := map[string]string{"first.user": "firstuser2", "second.user": "seconduser2", "explicit": "firstuser", "plain.order": "orderapi", "seconduser": "seconduser"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("reversed=%v got %v want %v", reversed, got, want)
		}
	}
}
