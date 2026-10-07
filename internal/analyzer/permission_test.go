package analyzer

import (
	"strings"
	"testing"

	"github.com/alecthomas/participle/v2"
	"go.yorun.ai/skel/internal/parser/grammar"
	"go.yorun.ai/skel/schema"
)

func TestPermissionRequireSupportsNestedExpressions(t *testing.T) {
	parser := participle.MustBuild[grammar.SkelContent](grammar.Options...)
	content, err := parser.ParseString("permission.skel", strings.TrimSpace(`
domain demo

resource User {
    @desc("lookup check")
    check byId {
        @desc("lookup input")
        @sensitive
        input {
            @sensitive
            userId: int
        }
    }

    action read
    action update {
        check self {
            input {
                userId: int
            }
        }
    }
    action manage
}

actor UserActor {
    via client {}
    auth {
        credential {
            session: string
        }
        info {
            userId: int
        }
    }
}

api service UserApiService { auth required
    for UserActor

    method updateProfile {
        require any(
            User:manage,
            all(
                User:read,
                User:update:byId(userId),
                User:update:self(userId)
            )
        )
        input {
            userId: int
        }
    }
}
`))
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}
	if err = content.Finalize(); err != nil {
		t.Fatalf("Finalize() error = %v", err)
	}
	if len(content.Entries) != 3 {
		t.Fatalf("unexpected entry count: %d", len(content.Entries))
	}
	domain := mustAnalyze(t, content).Schema()
	if len(domain.Services()) != 1 {
		t.Fatalf("unexpected service count: %d", len(domain.Services()))
	}
	checkService := domain.Resources()[0].CheckService
	checkMethod := checkService.Methods[0]
	if len(checkMethod.Arguments) != 2 {
		t.Fatalf("unexpected resource check argument count: %d", len(checkMethod.Arguments))
	}
	if checkMethod.Arguments[0].Name != "code" || checkMethod.Arguments[0].Source != schema.ArgumentSourcePermissionCode || checkMethod.Arguments[0].Type.Kind != schema.TypeKindScalar || checkMethod.Arguments[0].Type.Scalar != schema.ScalarString {
		t.Fatalf("unexpected resource check code argument: %+v", checkMethod.Arguments[0])
	}
	if !checkMethod.Arguments[1].Sensitive || !checkMethod.ArgumentsData.Members[1].Sensitive {
		t.Fatalf("resource check argument should preserve @sensitive: %+v", checkMethod.Arguments[1])
	}
	if checkMethod.InputDescription != "lookup input" {
		t.Fatalf("unexpected resource check input description: %q", checkMethod.InputDescription)
	}
	if !checkMethod.ArgumentsSensitive {
		t.Fatal("expected resource check input to be sensitive as a whole")
	}
	if checkMethod.Description != "lookup check" {
		t.Fatalf("unexpected resource check description: %q", checkMethod.Description)
	}

	require := domain.Services()[0].Methods[0].Require
	if require.Expression.Mode != schema.PermissionRequireModeAny {
		t.Fatalf("unexpected root mode: %s", require.Expression.Mode)
	}
	if len(require.Expression.Children) != 2 {
		t.Fatalf("unexpected root children: %d", len(require.Expression.Children))
	}
	if require.Expression.Children[0].Code != "demo.User:manage" {
		t.Fatalf("unexpected first code: %s", require.Expression.Children[0].Code)
	}

	allExpr := require.Expression.Children[1]
	if allExpr.Mode != schema.PermissionRequireModeAll || len(allExpr.Children) != 3 {
		t.Fatalf("unexpected all expr: %+v", allExpr)
	}
	if allExpr.Children[0].Code != "demo.User:read" {
		t.Fatalf("unexpected nested code: %s", allExpr.Children[0].Code)
	}

	checkAllExpr := allExpr.Children[1]
	if checkAllExpr.Mode != schema.PermissionRequireModeAll || len(checkAllExpr.Children) != 2 {
		t.Fatalf("unexpected resource check all expr: %+v", checkAllExpr)
	}
	if checkAllExpr.Children[0].Code != "demo.User:update" {
		t.Fatalf("unexpected resource check code: %s", checkAllExpr.Children[0].Code)
	}
	checkExpr := checkAllExpr.Children[1]
	if checkExpr.Mode != schema.PermissionRequireModeCheck || checkExpr.Check.MethodSkelName != "checkById" {
		t.Fatalf("unexpected resource check expr: %+v", checkExpr)
	}
	if len(checkExpr.Check.Arguments) != 1 || checkExpr.Check.Arguments[0].Name != "userId" {
		t.Fatalf("unexpected resource check invocation arguments: %+v", checkExpr.Check.Arguments)
	}

	checkAllExpr = allExpr.Children[2]
	if checkAllExpr.Mode != schema.PermissionRequireModeAll || len(checkAllExpr.Children) != 2 {
		t.Fatalf("unexpected action check all expr: %+v", checkAllExpr)
	}
	if checkAllExpr.Children[0].Code != "demo.User:update" {
		t.Fatalf("unexpected action check code: %s", checkAllExpr.Children[0].Code)
	}
	checkExpr = checkAllExpr.Children[1]
	if checkExpr.Mode != schema.PermissionRequireModeCheck || checkExpr.Check.MethodSkelName != "checkUpdateSelf" {
		t.Fatalf("unexpected action check expr: %+v", checkExpr)
	}
}

func TestResourceActionCheckCannotDuplicateResourceCheck(t *testing.T) {
	parser := participle.MustBuild[grammar.SkelContent](grammar.Options...)
	content, err := parser.ParseString("permission.skel", strings.TrimSpace(`
domain demo

resource User {
    check byId {
        input {
            id: int
        }
    }
    action delete {
        check byId {
            input {
                id: int
            }
        }
    }
}
`))
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}
	if err = content.Finalize(); err != nil {
		t.Fatalf("Finalize() error = %v", err)
	}
	expectAnalyzeDiagnosticsContains(t, "duplicated resource action check", content)
}

func TestPermissionRequireCheckSupportsSingleListWildcardPath(t *testing.T) {
	content := parseResourceTestContent(t, `
domain demo

data User {
    id: int
}

resource User {
    action update {
        check byIds {
            input {
                ids: list<int>
            }
        }
    }
}

api service UserApiService { auth required  for ClientActor
    method update {
        require User:update:byIds(users[*].id)
        input {
            users: list<User>
        }
    }
}

actor ClientActor { via client {} }
`)
	domain := mustAnalyze(t, content).Schema()
	require := domain.Services()[0].Methods[0].Require
	checkArgument := require.Expression.Children[1].Check.Arguments[0]
	if checkArgument.JsonPath != "users[*].id" {
		t.Fatalf("unexpected json path: %s", checkArgument.JsonPath)
	}
	if checkArgument.Type.Kind != schema.TypeKindList || checkArgument.Type.List.Element.Scalar != schema.ScalarInt {
		t.Fatalf("unexpected check argument type: %+v", checkArgument.Type)
	}
}

func TestPermissionRequireCheckRejectsMultipleListWildcardPaths(t *testing.T) {
	content := parseResourceTestContent(t, `
domain demo

data Item {
    id: int
}

data Order {
    items: list<Item>
}

resource Order {
    action update {
        check byItemIds {
            input {
                ids: list<int>
            }
        }
    }
}

api service OrderApiService { auth required  for ClientActor
    method update {
        require Order:update:byItemIds(orders[*].items[*].id)
        input {
            orders: list<Order>
        }
    }
}

actor ClientActor { via client {} }
`)
	expectAnalyzeDiagnosticsContains(t, "supports at most one [*]", content)
}

func TestPermissionRequireCheckRejectsTrailingListWildcardPath(t *testing.T) {
	content := parseResourceTestContent(t, `
domain demo

data User {
    id: int
}

resource User {
    action update {
        check byUsers {
            input {
                users: list<User>
            }
        }
    }
}

api service UserApiService { auth required  for ClientActor
    method update {
        require User:update:byUsers(users[*])
        input {
            users: list<User>
        }
    }
}

actor ClientActor { via client {} }
`)
	expectAnalyzeDiagnosticsContains(t, "cannot end with [*]", content)
}

func parseResourceTestContent(t *testing.T, content string) *grammar.SkelContent {
	t.Helper()
	parser := participle.MustBuild[grammar.SkelContent](grammar.Options...)
	parsed, err := parser.ParseString("resource.skel", strings.TrimSpace(content))
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}
	if err = parsed.Finalize(); err != nil {
		t.Fatalf("Finalize() error = %v", err)
	}
	return parsed
}
