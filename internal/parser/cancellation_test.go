package parser

import (
	"context"
	"errors"
	"github.com/alecthomas/participle/v2/lexer"
	"strings"
	"testing"
)

type countingSyntaxError struct{ calls int }

func (e *countingSyntaxError) Error() string   { return "input.skel:1:2: bad syntax" }
func (e *countingSyntaxError) Message() string { e.calls++; return "bad syntax" }
func (e *countingSyntaxError) Position() lexer.Position {
	return lexer.Position{Filename: "input.skel", Line: 1, Column: 2}
}
func TestSyntaxErrorBuildsMessageOnce(t *testing.T) {
	original := new(countingSyntaxError)
	err := normalizeSyntaxError(original, false)
	for range 10 {
		if err.Error() != original.Error() {
			t.Fatal("error text changed")
		}
	}
	if original.calls != 1 {
		t.Fatalf("built message %d times", original.calls)
	}
	if !errors.Is(err, original) {
		t.Fatal("error cause was lost")
	}
}

type cancellationContext struct {
	context.Context
	cancel    context.CancelFunc
	remaining int
}

func (c *cancellationContext) Err() error {
	c.remaining--
	if c.remaining == 0 {
		c.cancel()
	}
	return c.Context.Err()
}
func TestParserCancelsDuringTokenLoading(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	controlled := &cancellationContext{Context: ctx, cancel: cancel, remaining: 100}
	result, err := ParseSourceContext(controlled, "input.skel", []byte("domain demo\n"+strings.Repeat("data User {}\n", 1000)))
	if !errors.Is(err, context.Canceled) || result.Content != nil {
		t.Fatalf("cancelled parse returned syntax: %v %v", result, err)
	}
}
