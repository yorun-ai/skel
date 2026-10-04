package grammar

import (
	"unicode"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type AuthMarker struct {
	Pos   lexer.Position
	Value string
}

func (m *AuthMarker) Parse(lex *lexer.PeekingLexer) error {
	token := lex.Peek()
	if token.Value != "auth" && token.Value != "noauth" {
		return participle.NextMatch
	}
	lex.Next()
	m.Pos = token.Pos
	m.Value = token.Value
	if token.Value == "auth" {
		switch lex.Peek().Value {
		case "required", "optional", "anonymous", "off":
			m.Value = lex.Next().Value
		}
	}
	return nil
}

type TaskTrigger struct {
	Pos        lexer.Position
	Decorators []*Decorator `parser:"(@@ (Newline)*)*"`
	Name       *Identifier  `parser:"\"trigger\" @@"`
	Input      *MethodInput `parser:"\"{\" (Newline)* (@@ (Newline)*)? \"}\""`
}

type MethodInput struct {
	Pos        lexer.Position
	Decorators []*Decorator `parser:"(@@ (Newline)*)*"`
	Arguments  []*Argument  `parser:"\"input\" \"{\" (Newline)* (@@ (Newline)*)* \"}\""`
}

type MethodOutput struct {
	Pos        lexer.Position
	Decorators []*Decorator `parser:"(@@ (Newline)*)*"`
	Type       *Type        `parser:"\"output\" @@"`
}

type Argument struct {
	Pos        lexer.Position
	Decorators []*Decorator `parser:"(@@ (Newline)*)*"`
	Name       *Identifier  `parser:"@@"`
	Type       *Type        `parser:"\":\" @@"`
}

type Type struct {
	Pos       lexer.Position
	Plain     *PlainType     `parser:"(  @@"`
	List      *ListType      `parser:" | @@"`
	Map       *MapType       `parser:" | @@"`
	Reference *ReferenceType `parser:" | @@)"`
	Nullable  bool           `parser:"@\"?\"?"`
}

type PlainType string

var PlainTypes = []PlainType{
	Int,
	Float,
	Boolean,
	String,

	Decimal,
	Binary,

	Timestamp,
	Duration,
	LocalDate,
	LocalTime,
	LocalDateTime,

	UUID,
	JSON,
}

var plainTypeByName = map[string]PlainType{
	string(Int):     Int,
	string(Float):   Float,
	string(Boolean): Boolean,
	string(String):  String,

	string(Decimal): Decimal,
	string(Binary):  Binary,

	string(Timestamp):     Timestamp,
	string(Duration):      Duration,
	string(LocalDate):     LocalDate,
	string(LocalTime):     LocalTime,
	string(LocalDateTime): LocalDateTime,

	string(UUID): UUID,
	string(JSON): JSON,
}

const (
	Int     PlainType = "int"
	Float   PlainType = "float"
	Boolean PlainType = "bool"
	String  PlainType = "string"

	Decimal PlainType = "decimal"
	Binary  PlainType = "binary"

	Timestamp     PlainType = "timestamp"
	Duration      PlainType = "duration"
	LocalDate     PlainType = "localdate"
	LocalTime     PlainType = "localtime"
	LocalDateTime PlainType = "localdatetime"

	UUID PlainType = "uuid"
	JSON PlainType = "json"
)

func (s *PlainType) Parse(lex *lexer.PeekingLexer) error {
	token := lex.Peek()

	value, ok := plainTypeByName[token.Value]
	if !ok {
		return participle.NextMatch
	}

	lex.Next()
	*s = value
	return nil
}

type ListType struct {
	Pos   lexer.Position
	Value *Type `parser:"\"list\" \"<\" @@ \">\""`
}

type MapType struct {
	Pos   lexer.Position
	Key   *Type `parser:"\"map\" \"<\" @@"`
	Value *Type `parser:"\",\" @@ \">\""`
}

type ReferenceType struct {
	Pos           lexer.Position
	Name          *QualifiedName `parser:"@@"`
	TypeArguments []*Type        `parser:"(\"<\" (@@ (\",\")?)* \">\")?"`
}

// Identifier wrap Identifier for accurate position
type Identifier struct {
	Pos   lexer.Position
	Value string `parser:"@Identifier"`
}

func parseIdentifier(lex *lexer.PeekingLexer) (*Identifier, error) {
	token := lex.Next()
	if !isIdentifierToken(token) {
		return nil, participle.Errorf(token.Pos, "expected identifier")
	}
	return &Identifier{
		Pos:   token.Pos,
		Value: token.Value,
	}, nil
}

func isIdentifierToken(token *lexer.Token) bool {
	if token == nil || token.Value == "" {
		return false
	}
	first := rune(token.Value[0])
	if first != '_' && !unicode.IsLetter(first) {
		return false
	}
	for _, r := range token.Value[1:] {
		if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
