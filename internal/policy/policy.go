// Package policy owns authentication inheritance and permission composition.
// It operates on caller-owned expression types so schema and descriptor share
// language rules without depending on one another's representations.
package policy

import "fmt"

// NormalizeAuth interprets defaults and legacy spellings for one declaration.
func NormalizeAuth(mode, owner string) string {
	switch mode {
	case "", "unset":
		if owner == "method" {
			return "inherit"
		}
		return "required"
	case "inherit":
		if owner == "service" {
			return "required"
		}
		return mode
	case "auth":
		return "required"
	case "noauth":
		if owner == "web" {
			return "off"
		}
		return "optional"
	default:
		return mode
	}
}

// EffectiveAuth derives a method policy without evaluating a request.
func EffectiveAuth(service, method string) (string, error) {
	service = NormalizeAuth(service, "service")
	method = NormalizeAuth(method, "method")
	if service != "required" && service != "optional" && service != "anonymous" {
		return "", fmt.Errorf("unsupported service auth mode %q", service)
	}
	if method == "inherit" {
		return service, nil
	}
	if method != "required" && method != "optional" && method != "anonymous" {
		return "", fmt.Errorf("unsupported method auth mode %q", method)
	}
	return method, nil
}

// Expressions adapts permission trees without introducing a second model.
type Expressions[T any] struct {
	// Expand optionally expands an unresolved source term into code/check nodes.
	Expand   func(*T) *T
	Children func(*T) []*T
	Copy     func(*T, []*T) (*T, error)
	All      func([]*T) *T
}

// Clone copies an expression, rejecting cyclic or missing child nodes. Each
// representation owns leaf copying and retains its reference-resolution state.
func (ops Expressions[T]) Clone(root *T) (*T, error) {
	active := map[*T]bool{}
	var visit func(*T) (*T, error)
	visit = func(node *T) (*T, error) {
		if node == nil {
			return nil, fmt.Errorf("permission expression is nil")
		}
		if active[node] {
			return nil, fmt.Errorf("cyclic permission expression")
		}
		active[node] = true
		defer delete(active, node)
		value := node
		if ops.Expand != nil {
			value = ops.Expand(value)
		}
		var children []*T
		if original := ops.Children(value); original != nil {
			children = make([]*T, len(original))
			for index, child := range original {
				copied, err := visit(child)
				if err != nil {
					return nil, fmt.Errorf("children[%d]: %w", index, err)
				}
				children[index] = copied
			}
		}
		return ops.Copy(value, children)
	}
	return visit(root)
}

// Conjoin combines service and method requirements in declaration order. It
// preserves disjunctions, checks and argument bindings, with no request-time
// evaluation or reordering. The returned expression does not alias input nodes.
func (ops Expressions[T]) Conjoin(service, method *T) (*T, error) {
	var children []*T
	for index, root := range []*T{service, method} {
		if root == nil {
			continue
		}
		copied, err := ops.Clone(root)
		if err != nil {
			return nil, fmt.Errorf("%s.require: %w", []string{"service", "method"}[index], err)
		}
		children = append(children, copied)
	}
	switch len(children) {
	case 0:
		return nil, nil
	case 1:
		return children[0], nil
	default:
		return ops.All(children), nil
	}
}
