package analyzer

import (
	"testing"

	"go.yorun.ai/skel/internal/model"
	"go.yorun.ai/skel/internal/parser/grammar"
)

func TestAnalyzeReturnsErrorWhenDataReferencesConfig(t *testing.T) {
	expectAnalyzeDiagnosticsContains(t, "config SiteConfig cannot be used as a value type", &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Config: &grammar.Data{
					Name:      ident("SiteConfig"),
					Qualifier: ident("instant"),
					Members: []*grammar.DataMember{
						{Name: ident("title"), Type: plainType(grammar.String)},
					},
				},
			},
			{
				Data: &grammar.Data{
					Name: ident("Page"),
					Members: []*grammar.DataMember{
						{Name: ident("site"), Type: refGrammarType("SiteConfig")},
					},
				},
			},
		},
	})
}

func TestAnalyzeAllowsConfigReferencesData(t *testing.T) {
	mustAnalyze(t, &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Data: &grammar.Data{
					Name: ident("Database"),
					Members: []*grammar.DataMember{
						{Name: ident("host"), Type: plainType(grammar.String)},
					},
				},
			},
			{
				Config: &grammar.Data{
					Name:      ident("AppConfig"),
					Qualifier: ident("eternal"),
					Members: []*grammar.DataMember{
						{Name: ident("database"), Type: refGrammarType("Database")},
					},
				},
			},
		},
	})
}

func TestAnalyzeAllowsConfigReferencesEnum(t *testing.T) {
	domain := mustAnalyze(t, &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Enum: &grammar.Enum{
					Name: ident("UserStatus"),
					Items: []*grammar.EnumItem{
						{Name: ident("ACTIVE")},
					},
				},
			},
			{
				Config: &grammar.Data{
					Name:      ident("AppConfig"),
					Qualifier: ident("eternal"),
					Members: []*grammar.DataMember{
						{Name: ident("defaultStatus"), Type: refGrammarType("UserStatus")},
					},
				},
			},
		},
	}).Model()
	if domain.Configs()[0].Members[0].Type.Kind != model.TypeKindEnum {
		t.Fatalf("unexpected config member type: %v", domain.Configs()[0].Members[0].Type.Kind)
	}
}

func TestAnalyzeAllowsConfigListValueEnum(t *testing.T) {
	domain := mustAnalyze(t, &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Enum: &grammar.Enum{
					Name: ident("UserStatus"),
					Items: []*grammar.EnumItem{
						{Name: ident("ACTIVE")},
					},
				},
			},
			{
				Config: &grammar.Data{
					Name:      ident("AppConfig"),
					Qualifier: ident("eternal"),
					Members: []*grammar.DataMember{
						{Name: ident("statuses"), Type: listType(refGrammarType("UserStatus"))},
					},
				},
			},
		},
	}).Model()
	if domain.Configs()[0].Members[0].Type.List.Value.Kind != model.TypeKindEnum {
		t.Fatalf("unexpected list value type: %v", domain.Configs()[0].Members[0].Type.List.Value.Kind)
	}
}

func TestAnalyzeAllowsConfigMemberBinary(t *testing.T) {
	mustAnalyze(t, &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Config: &grammar.Data{
					Name:      ident("AppConfig"),
					Qualifier: ident("eternal"),
					Members: []*grammar.DataMember{
						{Name: ident("payload"), Type: plainType(grammar.Binary)},
					},
				},
			},
		},
	})
}

func TestAnalyzeAllowsConfigListBinary(t *testing.T) {
	mustAnalyze(t, &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Config: &grammar.Data{
					Name:      ident("AppConfig"),
					Qualifier: ident("eternal"),
					Members: []*grammar.DataMember{
						{Name: ident("payloads"), Type: listType(plainType(grammar.Binary))},
					},
				},
			},
		},
	})
}

func TestAnalyzeAllowsConfigListData(t *testing.T) {
	mustAnalyze(t, &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Data: &grammar.Data{
					Name: ident("Database"),
					Members: []*grammar.DataMember{
						{Name: ident("host"), Type: plainType(grammar.String)},
					},
				},
			},
			{
				Config: &grammar.Data{
					Name:      ident("AppConfig"),
					Qualifier: ident("eternal"),
					Members: []*grammar.DataMember{
						{Name: ident("databases"), Type: listType(refGrammarType("Database"))},
					},
				},
			},
		},
	})
}

func TestAnalyzeAllowsConfigMapData(t *testing.T) {
	mustAnalyze(t, &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Data: &grammar.Data{
					Name: ident("Database"),
					Members: []*grammar.DataMember{
						{Name: ident("host"), Type: plainType(grammar.String)},
					},
				},
			},
			{
				Config: &grammar.Data{
					Name:      ident("AppConfig"),
					Qualifier: ident("eternal"),
					Members: []*grammar.DataMember{
						{Name: ident("databases"), Type: mapType(plainType(grammar.String), refGrammarType("Database"))},
					},
				},
			},
		},
	})
}

func TestAnalyzeAllowsConfigMapBinary(t *testing.T) {
	mustAnalyze(t, &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Config: &grammar.Data{
					Name:      ident("AppConfig"),
					Qualifier: ident("eternal"),
					Members: []*grammar.DataMember{
						{Name: ident("payloads"), Type: mapType(plainType(grammar.String), plainType(grammar.Binary))},
					},
				},
			},
		},
	})
}

func TestAnalyzeAllowsConfigMapEnumKey(t *testing.T) {
	domain := mustAnalyze(t, &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Enum: &grammar.Enum{
					Name: ident("UserStatus"),
					Items: []*grammar.EnumItem{
						{Name: ident("ACTIVE")},
					},
				},
			},
			{
				Config: &grammar.Data{
					Name:      ident("AppConfig"),
					Qualifier: ident("eternal"),
					Members: []*grammar.DataMember{
						{Name: ident("statusNames"), Type: mapType(refGrammarType("UserStatus"), plainType(grammar.String))},
					},
				},
			},
		},
	}).Model()
	if domain.Configs()[0].Members[0].Type.Map.Key.Kind != model.TypeKindEnum {
		t.Fatalf("expected config map key enum")
	}
}

func TestAnalyzeAllowsConfigMapUUIDKey(t *testing.T) {
	domain := mustAnalyze(t, &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Config: &grammar.Data{
					Name:      ident("AppConfig"),
					Qualifier: ident("eternal"),
					Members: []*grammar.DataMember{
						{Name: ident("labelsById"), Type: mapType(plainType(grammar.UUID), plainType(grammar.String))},
					},
				},
			},
		},
	}).Model()
	key := domain.Configs()[0].Members[0].Type.Map.Key
	if key.Kind != model.TypeKindScalar || key.Scalar != model.ScalarUUID {
		t.Fatalf("unexpected config map key: %+v", key)
	}
}

func TestAnalyzeReturnsErrorForHardReferenceCycle(t *testing.T) {
	expectAnalyzeDiagnosticsContains(t, "hard reference chain detected", &grammar.SkelContent{
		Domain: domainContent("demo.user"),
		Entries: []*grammar.SkelEntry{
			{
				Data: &grammar.Data{
					Name: ident("User"),
					Members: []*grammar.DataMember{
						{Name: ident("profile"), Type: refGrammarType("Profile")},
					},
				},
			},
			{
				Data: &grammar.Data{
					Name: ident("Profile"),
					Members: []*grammar.DataMember{
						{Name: ident("user"), Type: refGrammarType("User")},
					},
				},
			},
		},
	})
}
