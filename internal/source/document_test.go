package source

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDocumentSharesIdentityDigestAndByteLocations(t *testing.T) {
	text := "α😀\r\nvalue: int\n"
	document := New("untitled:one", "", 12, text)
	require.Equal(t, sha256.Sum256([]byte(text)), document.Digest())
	require.Equal(t, "untitled:one", document.AnalysisPath())
	require.Equal(t, "α😀", document.Line(0))
	for _, offset := range []int{0, 2, 8, 13, len(text)} {
		line, column := document.Position(offset)
		require.Equal(t, offset, document.Offset(line, column))
	}
	require.Equal(t, Span{Start: 8, End: 13}, document.IdentifierSpan(2, 1, "value"))
}
