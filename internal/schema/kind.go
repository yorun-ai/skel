package schema

import (
	"fmt"
	"slices"
	"strings"
)

func ValidateKind(kind string) error {
	if slices.Contains(declarationKinds, DeclarationType(kind)) {
		return nil
	}
	values := make([]string, 0, len(declarationKinds))
	for _, declarationKind := range declarationKinds {
		values = append(values, string(declarationKind))
	}
	return fmt.Errorf("invalid schema declaration type %q, expected %s", kind, strings.Join(values, "/"))
}

// DeclarationTypes returns every supported top-level declaration type in
// stable schema order.
func DeclarationTypes() []DeclarationType {
	return append([]DeclarationType{}, declarationKinds...)
}
