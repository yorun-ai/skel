package typescript

import (
	"fmt"

	"go.yorun.ai/skel/schema"
)

type _WireMethod struct {
	Name            string
	ArgumentsSchema string
	ResultSchema    string
}

type _WireFactory struct {
	Code string
}

type _WireSchemaBuilder struct {
	data         map[*schema.Data]bool
	factoryNames map[*schema.Data]string
	types        []*schema.Type
	err          error
}

func (b *_WireSchemaBuilder) fail(format string, args ...any) {
	if b.err == nil {
		b.err = fmt.Errorf(format, args...)
	}
}

func newWireSchemaBuilder() *_WireSchemaBuilder {
	return &_WireSchemaBuilder{
		data:         map[*schema.Data]bool{},
		factoryNames: map[*schema.Data]string{},
		types:        make([]*schema.Type, 0),
	}
}

func methodArgumentsContainBinary(method *schema.Method) bool {
	for _, argument := range method.Arguments {
		if argument.Type.ContainsBinaryType() {
			return true
		}
	}
	return false
}

func methodResultContainsBinary(method *schema.Method) bool {
	return method.ResultType.ContainsBinaryType()
}
