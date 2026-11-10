package analyzer

import (
	"go.yorun.ai/skel/internal/parser"
	"testing"

	"go.yorun.ai/skel/internal/parser/grammar"
)

func TestParseActor(t *testing.T) {
	authSection := grammarActorAuthSection(
		[]*grammar.DataMember{{Name: ident("subject"), Type: plainType(grammar.String)}},
		[]*grammar.DataMember{{Name: ident("userId"), Type: plainType(grammar.Int)}},
	)
	authSection.Auth.Credential.Decorators = []*grammar.Decorator{{Name: ident("sensitive")}}
	authSection.Auth.Info.Decorators = []*grammar.Decorator{{Name: ident("sensitive")}}
	actor := parseActorTest(t, &grammar.Actor{
		Decorators: []*grammar.Decorator{
			{Name: ident("desc"), Value: decoratorValue(`"Portal admin"`)},
		},
		Pub:      true,
		Name:     ident("PortalAdminActor"),
		Sections: []*grammar.ActorSection{authSection},
	})
	if actor.Name != "PortalAdminActor" {
		t.Fatalf("unexpected actor name: %s", actor.Name)
	}
	if actor.SkelName != "" {
		t.Fatalf("unexpected actor skel name before domain normalize: %q", actor.SkelName)
	}
	if actor.Description != "Portal admin" {
		t.Fatalf("unexpected actor description: %q", actor.Description)
	}
	if !actor.Pub {
		t.Fatal("expected pub actor")
	}
	if actor.Auth.Credential == nil || actor.Auth.Credential.Name != "PortalAdminActorCredential" {
		t.Fatalf("unexpected actor credential: %+v", actor.Auth.Credential)
	}
	if !actor.Auth.Credential.Pub {
		t.Fatal("expected actor credential to follow actor pub")
	}
	if len(actor.Auth.Credential.Members) != 1 || actor.Auth.Credential.Members[0].Name != "subject" {
		t.Fatalf("unexpected actor credential members: %+v", actor.Auth.Credential.Members)
	}
	if !actor.Auth.Credential.Sensitive {
		t.Fatal("expected actor credential to be sensitive as a whole")
	}
	if actor.Auth.Info == nil || actor.Auth.Info.Name != "PortalAdminActorInfo" {
		t.Fatalf("unexpected actor info: %+v", actor.Auth.Info)
	}
	if !actor.Auth.Info.Pub {
		t.Fatal("expected actor info to follow actor pub")
	}
	if len(actor.Auth.Info.Members) != 1 || actor.Auth.Info.Members[0].Name != "userId" {
		t.Fatalf("unexpected actor info members: %+v", actor.Auth.Info.Members)
	}
	if !actor.Auth.Info.Sensitive {
		t.Fatal("expected actor info to be sensitive as a whole")
	}
}

func TestParseActorRejectsNonStringCredentialMember(t *testing.T) {
	expectActorDiagnostic(t, "actor credential member userId must be string or string?", &grammar.Actor{
		Name: ident("PortalAdminActor"),
		Sections: []*grammar.ActorSection{
			grammarActorAuthSection(
				[]*grammar.DataMember{{Name: ident("userId"), Type: plainType(grammar.Int)}},
				[]*grammar.DataMember{{Name: ident("userId"), Type: plainType(grammar.Int)}},
			),
		},
	})
}

func TestParseActorRejectsEmptyCredential(t *testing.T) {
	expectActorDiagnostic(t, "actor credential must have at least one member", &grammar.Actor{
		Name: ident("PortalAdminActor"),
		Sections: []*grammar.ActorSection{
			grammarActorAuthSection([]*grammar.DataMember{}, []*grammar.DataMember{{Name: ident("userId"), Type: plainType(grammar.Int)}}),
		},
	})
}

func TestParseActorAcceptsNullableCredentialMember(t *testing.T) {
	credentialType := plainType(grammar.String)
	credentialType.Nullable = true
	actor := parseActorTest(t, &grammar.Actor{
		Name: ident("PortalAdminActor"),
		Sections: []*grammar.ActorSection{
			grammarActorAuthSection(
				[]*grammar.DataMember{{Name: ident("subject"), Type: credentialType}, {Name: ident("token"), Type: plainType(grammar.String)}},
				[]*grammar.DataMember{{Name: ident("userId"), Type: plainType(grammar.Int)}},
			),
		},
	})

	if !actor.Auth.Credential.Members[0].Type.Nullable {
		t.Fatal("credential field lost its nullable type")
	}
}

func TestParseActorRejectsAllOptionalCredentialMembers(t *testing.T) {
	for _, fields := range []string{"token: string?", "token: string? session: string?"} {
		t.Run(fields, func(t *testing.T) {
			content, err := parser.ParseSource("actor.skel", []byte("domain demo\npub actor UserActor { auth { credential { "+fields+" } info { userId: string } } }"))
			if err != nil {
				t.Fatal(err)
			}
			expectActorDiagnostic(t, "actor credential must have at least one required string member", content.Entries[0].Actor)
		})
	}
}

func TestParseActorRejectsCredentialWithoutInfo(t *testing.T) {
	expectActorDiagnostic(t, "auth must define credential and info together", &grammar.Actor{
		Name: ident("PortalAdminActor"),
		Sections: []*grammar.ActorSection{
			grammarActorAuthSection([]*grammar.DataMember{{Name: ident("subject"), Type: plainType(grammar.String)}}, nil),
		},
	})
}

func TestParseActorRejectsInfoWithoutCredential(t *testing.T) {
	expectActorDiagnostic(t, "auth must define credential and info together", &grammar.Actor{
		Name: ident("PortalAdminActor"),
		Sections: []*grammar.ActorSection{
			grammarActorAuthSection(nil, []*grammar.DataMember{{Name: ident("userId"), Type: plainType(grammar.Int)}}),
		},
	})
}

func TestParseActorRejectsDuplicatedAuth(t *testing.T) {
	expectActorDiagnostic(t, "duplicated actor auth", &grammar.Actor{
		Name: ident("PortalAdminActor"),
		Sections: []*grammar.ActorSection{
			grammarActorAuthSection(
				[]*grammar.DataMember{{Name: ident("subject"), Type: plainType(grammar.String)}},
				[]*grammar.DataMember{{Name: ident("userId"), Type: plainType(grammar.Int)}},
			),
			grammarActorAuthSection(
				[]*grammar.DataMember{{Name: ident("tenant"), Type: plainType(grammar.String)}},
				[]*grammar.DataMember{{Name: ident("tenantId"), Type: plainType(grammar.Int)}},
			),
		},
	})
}

// TODO: Remove this compatibility test when legacy via syntax and warnings are removed.
func TestParseActorIgnoresVias(t *testing.T) {
	for _, names := range [][]string{nil, {"client", "client"}, {"partner", "UPPER"}} {
		reporter := newDiagnosticReporter()
		input := &grammar.Actor{Name: ident("PortalAdminActor")}
		for _, name := range names {
			input.Sections = append(input.Sections, new(grammar.ActorSection{Via: new(grammar.ActorVia{Name: ident(name)})}))
		}
		actor, valid := parseActor(reporter, input)
		if !valid || len(reporter.errors) != 0 || len(reporter.warnings) != len(names) {
			t.Fatalf("vias %v: valid=%v actor=%+v errors=%v warnings=%v", names, valid, actor, reporter.errors, reporter.warnings)
		}
	}
}

func grammarActorAuthSection(credentialMembers []*grammar.DataMember, infoMembers []*grammar.DataMember) *grammar.ActorSection {
	auth := &grammar.ActorAuth{}
	if credentialMembers != nil {
		auth.Credential = &grammar.ActorCredential{Members: credentialMembers}
	}
	if infoMembers != nil {
		auth.Info = &grammar.ActorInfo{Members: infoMembers}
	}
	return &grammar.ActorSection{Auth: auth}
}

func TestActorIdentifier(t *testing.T) {
	for _, tt := range []struct {
		name, body string
		valid      bool
	}{
		{"string", "@identifier id: string", true},
		{"uuid", "@identifier id: uuid", true},
		{"int", "@identifier id: int", true},
		{"optional marker", "id: string", true},
		{"nullable", "@identifier id: int?", false},
		{"float", "@identifier id: float", false},
		{"decimal", "@identifier id: decimal", false},
		{"list", "@identifier id: list<int>", false},
		{"argument", "@identifier(\"id\") id: int", false},
		{"duplicate", "@identifier @identifier id: int", false},
		{"two fields", "@identifier id: int @identifier other: string", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := "domain demo\npub actor UserActor { auth { credential { token: string } info { " + tt.body + " } } }"
			content, err := parser.ParseSource("actor.skel", []byte(source))
			if err != nil {
				t.Fatal(err)
			}
			reporter := newDiagnosticReporter()
			actor, valid := parseActor(reporter, content.Entries[0].Actor)
			if valid != tt.valid {
				t.Fatalf("valid=%v diagnostics=%v", valid, reporter.result())
			}
			if valid && tt.name != "optional marker" && actor.Auth.IdentifierField != "id" {
				t.Fatalf("identifier = %q", actor.Auth.IdentifierField)
			}
			again, validAgain := parseActor(newDiagnosticReporter(), content.Entries[0].Actor)
			if validAgain != valid || again.Auth.IdentifierField != actor.Auth.IdentifierField {
				t.Fatal("analysis mutated source")
			}
		})
	}
	for _, source := range []string{
		"domain demo\ndata User { @identifier id: int }",
		"domain demo\n@identifier actor UserActor {}",
		"domain demo\nactor UserActor { auth { credential { @identifier token: string } info {} } }",
	} {
		content, err := parser.ParseSource("invalid.skel", []byte(source))
		if err != nil {
			t.Fatal(err)
		}
		_, diagnostics := Analyze(content, nil)
		assertDiagnosticsContain(t, diagnostics, "unexpected decorator @identifier")
	}
}

func TestActorAuthenticationResultAndImpersonators(t *testing.T) {
	source := `domain demo
pub actor UserActor { auth {
    impersonator UserActor, OpsActor
    credential { session: string }
    info { @identifier userId: int }
} }
pub actor OpsActor { auth { credential { token: string } info { id: string } } }
`
	content, err := parser.ParseSource("actor.skel", []byte(source))
	if err != nil {
		t.Fatal(err)
	}
	analysis := mustAnalyze(t, content)
	auth := analysis.actorsMap["UserActor"].Auth
	if auth.Method.Name != "auth" || auth.Method.ResultType.Data != auth.Result || auth.Result.Name != "UserActorAuthResult" {
		t.Fatalf("wrong auth result: %+v", auth)
	}
	if len(auth.Result.Members) != 2 || auth.Result.Members[0].Name != "info" || auth.Result.Members[0].Type.Data != auth.Info || auth.Result.Members[0].Type.Nullable {
		t.Fatalf("wrong result members: %+v", auth.Result.Members)
	}
	if !auth.Result.Members[1].Type.Nullable || auth.Result.Members[1].Type.Data != auth.ImpersonatorInfo {
		t.Fatal("operator container must be nullable")
	}
	for i, name := range []string{"UserActor", "OpsActor"} {
		ref := auth.Impersonators[i]
		member := auth.ImpersonatorInfo.Members[i]
		if ref.ActorSkelName != "demo."+name || member.Name != ref.InfoFieldName || !member.Type.Nullable || member.Type.Data != analysis.actorsMap[name].Auth.Info {
			t.Fatalf("wrong operator mapping: %+v / %+v", ref, member)
		}
	}
	plain := analysis.actorsMap["OpsActor"].Auth
	if len(plain.Result.Members) != 1 || plain.ImpersonatorInfo != nil {
		t.Fatal("plain auth must only return info")
	}
}

func TestActorImpersonatorValidation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		clause string
		extra  string
		want   string
	}{
		{"undefined", "impersonator MissingActor", "", "undefined impersonator"},
		{"no auth", "impersonator OpsActor", "actor OpsActor {}", "must declare auth info"},
		{"duplicate", "impersonator UserActor, UserActor", "", "duplicated impersonator"},
		{"generated collision", "", "data UserActorAuthResult {}", "UserActorAuthResult"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := "domain demo\nactor UserActor { auth {\n" + tc.clause + "\ncredential { token: string } info { id: string } } }\n" + tc.extra
			content, err := parser.ParseSource("actor.skel", []byte(source))
			if err != nil {
				t.Fatal(err)
			}
			_, diagnostics := Analyze(content, nil)
			assertDiagnosticsContain(t, diagnostics, tc.want)
		})
	}
}

func TestActorImpersonatorImportedReference(t *testing.T) {
	source, err := parser.ParseSource("shared.skel", []byte(`domain shared
pub actor OpsActor { auth { credential { token: string } info { id: string } } }`))
	if err != nil {
		t.Fatal(err)
	}
	imported := mustAnalyze(t, source)
	source, err = parser.ParseSource("app.skel", []byte(`domain app
import shared as ops
pub actor UserActor { auth { impersonator ops.OpsActor credential { token: string } info { id: string } } }`))
	if err != nil {
		t.Fatal(err)
	}
	analysis := mustAnalyze(t, source, imported)
	auth := analysis.actorsMap["UserActor"].Auth
	kind := auth.ImpersonatorInfo.Members[0].Type
	if kind.ExternalDomain != "shared" || kind.ExternalAlias != "ops" || !kind.ExternalAliasExplicit || auth.Impersonators[0].ActorSkelName != "shared.OpsActor" {
		t.Fatalf("lost imported identity: %+v", kind)
	}
}

func TestActorImpersonatorNameCollisionAndVisibility(t *testing.T) {
	for _, tc := range []struct {
		name     string
		imported string
		clause   string
		local    string
		want     string
	}{
		{"field collision", "pub actor OpsActor", "shared.OpsActor, OpsActor", "pub actor OpsActor { auth { credential { token: string } info {} } }", "info field opsActorInfo conflicts"},
		{"private import", "actor OpsActor", "shared.OpsActor", "", "is not public"},
		{"private local", "pub actor OpsActor", "LocalActor", "actor LocalActor { auth { credential { token: string } info {} } }", "cannot expose private impersonator"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			content, err := parser.ParseSource("shared.skel", []byte("domain shared\n"+tc.imported+" { auth { credential { token: string } info {} } }"))
			if err != nil {
				t.Fatal(err)
			}
			imported := mustAnalyze(t, content)
			content, err = parser.ParseSource("app.skel", []byte("domain app\nimport shared\npub actor UserActor { auth { impersonator "+tc.clause+"\ncredential { token: string } info {} } }\n"+tc.local))
			if err != nil {
				t.Fatal(err)
			}
			_, diagnostics := Analyze(content, []*Analysis{imported})
			assertDiagnosticsContain(t, diagnostics, tc.want)
		})
	}
}
