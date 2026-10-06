package codegen

import "go.yorun.ai/skel/schema"

// TypeVisitor observes one type during a pre-order walk. Returning an error
// stops the walk immediately.
type TypeVisitor func(*schema.Type) error

// WalkType visits a type and its structural children: list elements, map keys
// and values, and generic type arguments. Shared type nodes are visited once.
func WalkType(type_ *schema.Type, visit TypeVisitor) error {
	return WalkTypes([]*schema.Type{type_}, visit)
}

// WalkTypes visits several roots while sharing traversal state. A type node
// referenced by more than one root is visited once.
func WalkTypes(types []*schema.Type, visit TypeVisitor) error {
	seenTypes := map[*schema.Type]bool{}
	for _, kind := range types {
		if err := walkType(kind, visit, false, seenTypes, nil); err != nil {
			return err
		}
	}
	return nil
}

// VisitType is the non-failing form of WalkType.
func VisitType(type_ *schema.Type, visit func(*schema.Type)) {
	VisitTypes([]*schema.Type{type_}, visit)
}

// VisitTypes is the non-failing form of WalkTypes.
func VisitTypes(types []*schema.Type, visit func(*schema.Type)) {
	_ = WalkTypes(types, func(kind *schema.Type) error {
		visit(kind)
		return nil
	})
}

// WalkTypeGraph additionally follows members of referenced data declarations.
// It is intended for graph-wide questions such as wire-schema discovery and
// safely terminates on recursive data definitions.
func WalkTypeGraph(type_ *schema.Type, visit TypeVisitor) error {
	return WalkTypeGraphs([]*schema.Type{type_}, visit)
}

// WalkTypeGraphs follows referenced data from several roots while sharing
// traversal state across the complete graph.
func WalkTypeGraphs(types []*schema.Type, visit TypeVisitor) error {
	seenTypes := map[*schema.Type]bool{}
	seenData := map[*schema.Data]bool{}
	for _, kind := range types {
		if err := walkType(kind, visit, true, seenTypes, seenData); err != nil {
			return err
		}
	}
	return nil
}

// VisitTypeGraphs is the non-failing form of WalkTypeGraphs.
func VisitTypeGraphs(types []*schema.Type, visit func(*schema.Type)) {
	_ = WalkTypeGraphs(types, func(kind *schema.Type) error {
		visit(kind)
		return nil
	})
}

func walkType(type_ *schema.Type, visit TypeVisitor, followData bool, seenTypes map[*schema.Type]bool, seenData map[*schema.Data]bool) error {
	if type_ == nil || seenTypes[type_] {
		return nil
	}
	seenTypes[type_] = true
	if err := visit(type_); err != nil {
		return err
	}
	switch type_.Kind {
	case schema.TypeKindList:
		if type_.List != nil {
			return walkType(type_.List.Value, visit, followData, seenTypes, seenData)
		}
	case schema.TypeKindMap:
		if type_.Map != nil {
			if err := walkType(type_.Map.Key, visit, followData, seenTypes, seenData); err != nil {
				return err
			}
			return walkType(type_.Map.Value, visit, followData, seenTypes, seenData)
		}
	case schema.TypeKindData:
		for _, argument := range type_.TypeArguments {
			if err := walkType(argument, visit, followData, seenTypes, seenData); err != nil {
				return err
			}
		}
		if followData && type_.Data != nil && !seenData[type_.Data] {
			seenData[type_.Data] = true
			for _, member := range type_.Data.Members {
				if member != nil {
					if err := walkType(member.Type, visit, followData, seenTypes, seenData); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
