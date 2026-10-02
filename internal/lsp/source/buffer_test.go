package source

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"
)

func TestBufferConvertsUTF16Positions(t *testing.T) {
	buffer := New("data 𐐀User {}\r\n")
	position := buffer.Position(len("data 𐐀"))
	assert.Equal(t, protocol.Position{Line: 0, Character: 7}, position)
	assert.Equal(t, len("data 𐐀"), buffer.Offset(position))
	assert.Equal(t, protocol.Range{
		Start: protocol.Position{Line: 0, Character: 7},
		End:   protocol.Position{Line: 0, Character: 11},
	}, buffer.IdentifierRange(1, 7, "User"))
}

func TestBufferScansIdentifiersOutsideCommentsAndStrings(t *testing.T) {
	buffer := New("data User { // Ignored\n value: Other @desc(\"Hidden\")\n}\n")
	tokens := buffer.IdentifierTokens()
	values := make([]string, 0, len(tokens))
	for _, token := range tokens {
		values = append(values, token.Value)
	}
	assert.Equal(t, []string{"data", "User", "{", "value", "Other", "desc", "}"}, values)
	assert.True(t, buffer.InNonCode(protocol.Position{Line: 0, Character: 20}))
	assert.True(t, buffer.InNonCode(protocol.Position{Line: 1, Character: 22}))
	assert.False(t, buffer.InNonCode(protocol.Position{Line: 1, Character: 2}))
}

func TestMountPathsAreNotSymbolReferences(t *testing.T) {
	text := "mount /ClientActor/other.User\nfor ClientActor"
	buffer := New(text)
	tokens := buffer.IdentifierTokens()
	if len(tokens) != 3 || tokens[0].Value != "mount" || tokens[1].Value != "for" || tokens[2].Value != "ClientActor" {
		t.Fatalf("path leaked into symbol tokens: %+v", tokens)
	}
	if !buffer.InNonCode(buffer.Position(12)) || buffer.InNonCode(buffer.Position(len(text))) {
		t.Fatal("incorrect code classification around mount path")
	}
}

func TestIndexedPositionsRoundTripAcrossLines(t *testing.T) {
	for _, text := range []string{"", "a\n", "a\r\nb𐐀c\n", "\n\n尾"} {
		buffer := New(text)
		for offset := range len(text) + 1 {
			if offset < len(text) && text[offset]&0xc0 == 0x80 {
				continue
			}
			assert.Equal(t, offset, buffer.Offset(buffer.Position(offset)), "text=%q offset=%d", text, offset)
		}
		assert.Equal(t, len(text), buffer.Offset(protocol.Position{Line: 100, Character: 100}))
	}
	assert.Equal(t, protocol.Position{}, (Buffer{}).Position(0))
}
