package analyzer

import (
	"testing"

	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/parser/grammar"
)

func TestParseService(t *testing.T) {
	service := parseServiceTest(t, &grammar.Service{
		Decorators: []*grammar.Decorator{
			{Name: ident("desc"), Value: decoratorValue(`"Client user service"`)},
		},
		Pub:       true,
		Name:      ident("UserService"),
		Audiences: []*grammar.ServiceAudience{serviceAllow("ClientActor", "client"), serviceAllow("OpenAPIActor")},
		Methods: []*grammar.Method{
			{
				Decorators: []*grammar.Decorator{
					{Name: ident("desc"), Value: decoratorValue(`"Get a user by ID"`)},
				},
				Name: ident("getUser"),
				Input: &grammar.MethodInput{
					Decorators: []*grammar.Decorator{
						{Name: ident("desc"), Value: decoratorValue(`"Request parameters"`)},
						{Name: ident("sensitive")},
					},
					Arguments: []*grammar.Argument{
						{
							Decorators: []*grammar.Decorator{
								{Name: ident("desc"), Value: decoratorValue(`"User ID"`)},
								{Name: ident("example"), Value: decoratorValue(`"10001"`)},
								{Name: ident("sensitive")},
							},
							Name: ident("userId"),
							Type: plainType(grammar.Int),
						},
					},
				},
				Output: &grammar.MethodOutput{
					Decorators: []*grammar.Decorator{
						{Name: ident("desc"), Value: decoratorValue(`"User information"`)},
						{Name: ident("example"), Value: decoratorValue(`{"id":10001}`)},
						{Name: ident("sensitive")},
					},
					Type: nullableType(refGrammarType("User")),
				},
			},
		},
	})

	if service.Name != "UserService" {
		t.Fatalf("unexpected service name: %s", service.Name)
	}
	if len(service.Audiences) != 2 {
		t.Fatalf("unexpected service for count: %d", len(service.Audiences))
	}
	if service.Description != "Client user service" {
		t.Fatalf("unexpected service description: %q", service.Description)
	}
	if !service.Pub {
		t.Fatal("expected pub service")
	}
	if service.Audiences[0].Actor != "ClientActor" || service.Audiences[1].Actor != "OpenAPIActor" {
		t.Fatalf("unexpected service audiences: %+v", service.Audiences)
	}
	if service.Audiences[0].Via != "client" {
		t.Fatalf("unexpected service for via: %+v", service.Audiences)
	}
	if len(service.Methods) != 1 {
		t.Fatalf("unexpected service method count: %d", len(service.Methods))
	}
	if service.Methods[0].Description != "Get a user by ID" {
		t.Fatalf("unexpected method description: %q", service.Methods[0].Description)
	}
	if service.Methods[0].InputDescription != "Request parameters" {
		t.Fatalf("unexpected input description: %q", service.Methods[0].InputDescription)
	}
	if service.Methods[0].OutputDescription != "User information" {
		t.Fatalf("unexpected output description: %q", service.Methods[0].OutputDescription)
	}
	if service.Methods[0].OutputExample != `{"id":10001}` {
		t.Fatalf("unexpected output example: %q", service.Methods[0].OutputExample)
	}
	if !service.Methods[0].ArgumentsSensitive || !service.Methods[0].ResultSensitive {
		t.Fatal("expected whole input and output sensitive metadata")
	}
	if service.Methods[0].Arguments[0].Description != "User ID" || service.Methods[0].Arguments[0].Example != `"10001"` {
		t.Fatalf("unexpected argument annotations: %+v", service.Methods[0].Arguments[0])
	}
	if !service.Methods[0].Arguments[0].Sensitive || !service.Methods[0].ArgumentsData.Members[0].Sensitive {
		t.Fatal("expected sensitive metadata on argument and generated argument data")
	}
	if service.Methods[0].ArgumentsData == nil || service.Methods[0].ArgumentsData.Name != "UserServiceGetUserArguments" {
		t.Fatalf("unexpected arguments data: %+v", service.Methods[0].ArgumentsData)
	}
}

func TestParseServiceSupportsTripleQuotedDescription(t *testing.T) {
	service := parseServiceTest(t, &grammar.Service{
		Decorators: []*grammar.Decorator{
			{Name: ident("desc"), Value: decoratorValue("\"\"\"\n    User service\n    Second line description\n\"\"\"")},
		},
		Name:      ident("UserService"),
		Audiences: []*grammar.ServiceAudience{serviceAllow("ClientActor")},
		Methods: []*grammar.Method{
			{
				Decorators: []*grammar.Decorator{
					{Name: ident("desc"), Value: decoratorValue("\"\"\"\n    Get a user by ID\n    Second line description\n\"\"\"")},
				},
				Name: ident("getUser"),
				Output: &grammar.MethodOutput{
					Type: plainType(grammar.String),
				},
			},
		},
	})

	if service.Description != "User service\nSecond line description" {
		t.Fatalf("unexpected service description: %q", service.Description)
	}
	if service.Methods[0].Description != "Get a user by ID\nSecond line description" {
		t.Fatalf("unexpected method description: %q", service.Methods[0].Description)
	}
}

func TestParseServiceAudiencesMissingActors(t *testing.T) {
	service := parseServiceTest(t, &grammar.Service{
		Name: ident("UserService"),
		Methods: []*grammar.Method{
			{
				Name:   ident("ping"),
				Output: &grammar.MethodOutput{Type: plainType(grammar.String)},
			},
		},
	})
	if len(service.Audiences) != 0 {
		t.Fatalf("expected empty audiences, got %+v", service.Audiences)
	}
}

func TestParseServiceAudiencesPub(t *testing.T) {
	service := parseServiceTest(t, &grammar.Service{
		Pub:  true,
		Name: ident("UserService"),
		Methods: []*grammar.Method{
			{
				Name:   ident("ping"),
				Output: &grammar.MethodOutput{Type: plainType(grammar.String)},
			},
		},
	})
	if !service.Pub {
		t.Fatal("expected pub service")
	}
}

func TestParseServiceSectionsAllowAnyOrderAndMethodAuthOverride(t *testing.T) {
	service := parseServiceTest(t, &grammar.Service{
		Name: ident("UserService"),
		Sections: []*grammar.ServiceSection{
			{
				Method: &grammar.Method{
					Name: ident("list"),
				},
			},
			{
				Auth: &grammar.AuthMarker{Value: "noauth"},
			},
			{
				Audience: serviceAllow("ClientActor", "client"),
			},
			{
				Method: &grammar.Method{
					Name: ident("update"),
					Auth: &grammar.AuthMarker{Value: "auth"},
				},
			},
		},
	})

	if service.Auth != model.AuthModeNoAuth {
		t.Fatalf("unexpected service auth: %s", service.Auth)
	}
	if len(service.Audiences) != 1 || service.Audiences[0].Actor != "ClientActor" || service.Audiences[0].Via != "client" {
		t.Fatalf("unexpected audiences: %+v", service.Audiences)
	}
	if service.Methods[0].Auth != model.AuthModeUnset {
		t.Fatalf("expected list auth to stay unset, got %s", service.Methods[0].Auth)
	}
	if service.Methods[1].Auth != model.AuthModeAuth {
		t.Fatalf("expected update to override auth, got %s", service.Methods[1].Auth)
	}
}

func TestParseServiceDefaultsAuthToUnset(t *testing.T) {
	service := parseServiceTest(t, &grammar.Service{
		Name: ident("UserService"),
		Methods: []*grammar.Method{
			{Name: ident("ping")},
		},
	})

	if service.Auth != model.AuthModeUnset {
		t.Fatalf("expected service auth unset, got %s", service.Auth)
	}
	if service.Methods[0].Auth != model.AuthModeUnset {
		t.Fatalf("expected method auth unset, got %s", service.Methods[0].Auth)
	}
}

func TestParseServiceRejectsExampleDecorator(t *testing.T) {
	expectServiceDiagnostic(t, "unexpected decorator @example", &grammar.Service{
		Decorators: []*grammar.Decorator{
			{Name: ident("example"), Value: decoratorValue(`"demo"`)},
		},
		Name:      ident("UserService"),
		Audiences: []*grammar.ServiceAudience{serviceAllow("ClientActor")},
		Methods: []*grammar.Method{
			{
				Name:   ident("getUser"),
				Output: &grammar.MethodOutput{Type: plainType(grammar.Int)},
			},
		},
	})
}

func TestParseServiceRejectsMethodExampleDecorator(t *testing.T) {
	expectServiceDiagnostic(t, "unexpected decorator @example", &grammar.Service{
		Name:      ident("UserService"),
		Audiences: []*grammar.ServiceAudience{serviceAllow("ClientActor")},
		Methods: []*grammar.Method{
			{
				Decorators: []*grammar.Decorator{
					{Name: ident("example"), Value: decoratorValue(`"demo"`)},
				},
				Name:   ident("getUser"),
				Output: &grammar.MethodOutput{Type: plainType(grammar.Int)},
			},
		},
	})
}

func TestParseServiceRejectsInputExampleDecorator(t *testing.T) {
	expectServiceDiagnostic(t, "unexpected decorator @example", &grammar.Service{
		Name:      ident("UserService"),
		Audiences: []*grammar.ServiceAudience{serviceAllow("ClientActor")},
		Methods: []*grammar.Method{
			{
				Name: ident("getUser"),
				Input: &grammar.MethodInput{
					Decorators: []*grammar.Decorator{
						{Name: ident("example"), Value: decoratorValue(`{"userId":10001}`)},
					},
					Arguments: []*grammar.Argument{
						{Name: ident("userId"), Type: plainType(grammar.Int)},
					},
				},
			},
		},
	})
}

func TestParseServiceReturnsErrorForDuplicatedMethod(t *testing.T) {
	expectServiceDiagnostic(t, "duplicated method getUser", &grammar.Service{
		Name:      ident("UserService"),
		Audiences: []*grammar.ServiceAudience{serviceAllow("ClientActor")},
		Methods: []*grammar.Method{
			{Name: ident("getUser")},
			{Name: ident("getUser")},
		},
	})
}

func TestParseServiceAudiencesArgumentNameMatchingMethodName(t *testing.T) {
	service := parseServiceTest(t, &grammar.Service{
		Name:      ident("UserService"),
		Audiences: []*grammar.ServiceAudience{serviceAllow("ClientActor")},
		Methods: []*grammar.Method{
			{
				Name: ident("getUser"),
				Input: &grammar.MethodInput{
					Arguments: []*grammar.Argument{
						{Name: ident("getUser"), Type: plainType(grammar.Int)},
					},
				},
			},
		},
	})
	if len(service.Methods) != 1 || len(service.Methods[0].Arguments) != 1 {
		t.Fatalf("expected parsed method argument to be preserved, got %+v", service.Methods)
	}
}

func TestParseServiceReturnsErrorWhenExampleHasNoDescription(t *testing.T) {
	expectServiceDiagnostic(t, "decorator @example must be used with @desc", &grammar.Service{
		Name:      ident("UserService"),
		Audiences: []*grammar.ServiceAudience{serviceAllow("ClientActor")},
		Methods: []*grammar.Method{
			{
				Name: ident("getUser"),
				Input: &grammar.MethodInput{
					Arguments: []*grammar.Argument{
						{
							Decorators: []*grammar.Decorator{
								{Name: ident("example"), Value: decoratorValue(`"10001"`)},
							},
							Name: ident("userId"),
							Type: plainType(grammar.Int),
						},
					},
				},
			},
		},
	})
}
