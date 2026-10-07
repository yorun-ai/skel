package tag_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	skeltag "go.yorun.ai/skel/tag"
)

func TestIsSensitive(t *testing.T) {
	for _, value := range []string{"sensitive", "index(0),sensitive", " sensitive , unknownFlag "} {
		require.True(t, skeltag.IsSensitive(reflect.StructTag(`skel:"`+value+`"`)))
	}
	for _, value := range []string{"", "notSensitive", "sensitiveExtra", "sensitive(false)"} {
		require.False(t, skeltag.IsSensitive(reflect.StructTag(`skel:"`+value+`"`)))
	}
}

func TestIsIdentifier(t *testing.T) {
	for _, value := range []string{"identifier", "sensitive,identifier", " identifier , unknownFlag "} {
		require.True(t, skeltag.IsIdentifier(reflect.StructTag(`skel:"`+value+`"`)))
	}
	for _, value := range []string{"", "notIdentifier", "identifierExtra", "identifier(false)"} {
		require.False(t, skeltag.IsIdentifier(reflect.StructTag(`skel:"`+value+`"`)))
	}
}

func TestIndex(t *testing.T) {
	index, found, err := skeltag.Index(`skel:" sensitive , index(2) "`)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, 2, index)
	_, found, err = skeltag.Index(`skel:"sensitive"`)
	require.NoError(t, err)
	require.False(t, found)
	for _, value := range []string{"index", "index()", "index(x)", "index(0", "index(0)extra", "index(0),index(1)", "index(0),index(0)", "index(999999999999999999999999)"} {
		t.Run(value, func(t *testing.T) {
			_, _, err := skeltag.Index(reflect.StructTag(`skel:"` + value + `"`))
			require.Error(t, err)
		})
	}
}
