package codegen

import (
	"fmt"

	"go.yorun.ai/skel/internal/model"
)

// InstantiateMembers substitutes a resolved data reference's type arguments in
// its members. It copies member/type expressions and preserves named declaration
// identities. It does not recursively expand named data, so recursive definitions
// terminate. Input must belong to a prepared semantic graph.
func InstantiateMembers(kind *model.Type) ([]*model.DataMember, error) {
	if kind == nil || kind.Kind != model.TypeKindData || kind.Data == nil {
		return nil, fmt.Errorf("expected a resolved data reference")
	}
	if err := validateModelType(kind); err != nil {
		return nil, err
	}
	bindings := map[*model.TypeParameter]*model.Type{}
	for i, p := range kind.Data.TypeParameters {
		bindings[p] = kind.TypeArguments[i]
	}
	seen := map[*model.Type]*model.Type{}
	argumentCopies := map[*model.Type]*model.Type{}
	var substitute func(*model.Type) *model.Type
	substitute = func(t *model.Type) *model.Type {
		if t == nil {
			return nil
		}
		if copy, ok := seen[t]; ok {
			return copy
		}
		copy := new(*t)
		seen[t] = copy
		if t.Kind == model.TypeKindTypeParameter {
			if arg := bindings[t.TypeParameter]; arg != nil {
				*copy = *copyTypeExpression(arg, argumentCopies)
				copy.Nullable = arg.Nullable || t.Nullable
				return copy
			}
		}
		if t.List != nil {
			copy.List = new(model.ListType{Value: substitute(t.List.Value)})
		}
		if t.Map != nil {
			copy.Map = new(model.MapType{Key: substitute(t.Map.Key), Value: substitute(t.Map.Value)})
		}
		copy.TypeArguments = nil
		for _, arg := range t.TypeArguments {
			copy.TypeArguments = append(copy.TypeArguments, substitute(arg))
		}
		return copy
	}
	members := make([]*model.DataMember, 0, len(kind.Data.Members))
	for _, m := range kind.Data.Members {
		copy := new(*m)
		copy.Type = substitute(m.Type)
		members = append(members, copy)
	}
	return members, nil
}

// Actual arguments belong to the caller's generic scope; copy their expressions
// without substituting the callee's parameters again.
func copyTypeExpression(kind *model.Type, seen map[*model.Type]*model.Type) *model.Type {
	if kind == nil {
		return nil
	}
	if copy, ok := seen[kind]; ok {
		return copy
	}
	copy := new(*kind)
	seen[kind] = copy
	if kind.List != nil {
		copy.List = new(model.ListType{Value: copyTypeExpression(kind.List.Value, seen)})
	}
	if kind.Map != nil {
		copy.Map = new(model.MapType{Key: copyTypeExpression(kind.Map.Key, seen), Value: copyTypeExpression(kind.Map.Value, seen)})
	}
	copy.TypeArguments = nil
	for _, arg := range kind.TypeArguments {
		copy.TypeArguments = append(copy.TypeArguments, copyTypeExpression(arg, seen))
	}
	return copy
}
