package hasher

import (
	"testing"

	"go.yorun.ai/skelc/internal/model"
)

func TestFillHashesPropagatesDataChangesToService(t *testing.T) {
	oldDomain := newHashTestDomain(t, "User service")
	newDomain := newHashTestDomain(t, "User service")
	newDomain.Data()[0].Members = append(newDomain.Data()[0].Members, &model.DataMember{
		Name: "nickname",
		Type: &model.Type{
			Kind:   model.TypeKindScalar,
			Scalar: model.ScalarString,
		},
	})

	fillHashes(t, oldDomain, newDomain)

	if oldDomain.Data()[0].Hash == newDomain.Data()[0].Hash {
		t.Fatal("expected data hash to change")
	}
	if oldDomain.Services()[0].Methods[0].Hash == newDomain.Services()[0].Methods[0].Hash {
		t.Fatal("expected method hash to change")
	}
	if oldDomain.Services()[0].Hash == newDomain.Services()[0].Hash {
		t.Fatal("expected service hash to change")
	}
	if oldDomain.Hash() == newDomain.Hash() {
		t.Fatal("expected domain hash to change")
	}
}

func TestFillHashesIncludesAllowVia(t *testing.T) {
	clientDomain := newHashAllowViaTestDomain(t, "client")
	openapiDomain := newHashAllowViaTestDomain(t, "openapi")

	fillHashes(t, clientDomain, openapiDomain)

	if clientDomain.Services()[0].Hash == openapiDomain.Services()[0].Hash {
		t.Fatal("expected service hash to change when for via changes")
	}
	if clientDomain.Webs()[0].Hash == openapiDomain.Webs()[0].Hash {
		t.Fatal("expected web hash to change when for via changes")
	}
	if clientDomain.Hash() == openapiDomain.Hash() {
		t.Fatal("expected domain hash to change when for via changes")
	}
}

func TestFillHashesIncludesSensitiveMetadata(t *testing.T) {
	oldDomain := newHashTestDomain(t, "User service")
	newDomain := newHashTestDomain(t, "User service")
	newDomain.Data()[0].Members[0].Sensitive = true
	newDomain.Services()[0].Methods[0].Arguments[0].Sensitive = true

	fillHashes(t, oldDomain, newDomain)

	if oldDomain.Data()[0].Hash == newDomain.Data()[0].Hash {
		t.Fatal("expected data hash to change when sensitive metadata changes")
	}
	if oldDomain.Services()[0].Methods[0].Hash == newDomain.Services()[0].Methods[0].Hash {
		t.Fatal("expected method hash to change when sensitive metadata changes")
	}
}

func TestFillHashesIncludesDeprecatedMetadata(t *testing.T) {
	oldDomain := newHashTestDomain(t, "User service")
	newDomain := newHashTestDomain(t, "User service")
	newDomain.Data()[0].Members[0].Deprecated = true
	newDomain.Data()[0].Members[0].DeprecatedReason = "Use id instead"
	newDomain.Services()[0].Methods[0].Deprecated = true
	newDomain.Services()[0].Methods[0].DeprecatedReason = "Use getProfile instead"
	newDomain.Services()[0].Methods[0].Arguments[0].Deprecated = true
	newDomain.Services()[0].Methods[0].Arguments[0].DeprecatedReason = "Use subject instead"

	fillHashes(t, oldDomain, newDomain)

	if oldDomain.Data()[0].Hash == newDomain.Data()[0].Hash {
		t.Fatal("expected data hash to change when member deprecation changes")
	}
	if oldDomain.Services()[0].Methods[0].Hash == newDomain.Services()[0].Methods[0].Hash {
		t.Fatal("expected method hash to change when deprecation changes")
	}
	if oldDomain.Services()[0].Hash == newDomain.Services()[0].Hash {
		t.Fatal("expected service hash to change when method deprecation changes")
	}
	if oldDomain.Hash() == newDomain.Hash() {
		t.Fatal("expected domain hash to change when deprecation changes")
	}
}

func TestFillHashesIncludesWholeSensitiveMetadata(t *testing.T) {
	t.Run("data", func(t *testing.T) {
		oldDomain := newHashTestDomain(t, "User service")
		newDomain := newHashTestDomain(t, "User service")
		newDomain.Data()[0].Sensitive = true

		fillHashes(t, oldDomain, newDomain)

		if oldDomain.Data()[0].Hash == newDomain.Data()[0].Hash {
			t.Fatal("expected data hash to change when whole-data sensitive metadata changes")
		}
	})

	for _, test := range []struct {
		name  string
		build func() (*model.Domain, *model.Data)
	}{
		{
			name: "config",
			build: func() (*model.Domain, *model.Data) {
				return newHashDataKindTestDomain(model.DataKindConfig)
			},
		},
		{
			name: "event payload",
			build: func() (*model.Domain, *model.Data) {
				return newHashDataKindTestDomain(model.DataKindEvent)
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			oldDomain, oldData := test.build()
			newDomain, newData := test.build()
			newData.Sensitive = true

			fillHashes(t, oldDomain, newDomain)

			if oldData.Hash == newData.Hash {
				t.Fatalf("expected %s hash to change when whole-value sensitive metadata changes", test.name)
			}
			if oldDomain.Hash() == newDomain.Hash() {
				t.Fatalf("expected domain hash to change when %s sensitive metadata changes", test.name)
			}
		})
	}

	for _, test := range []struct {
		name       string
		selectData func(*model.Actor) *model.Data
	}{
		{name: "actor credential", selectData: func(actor *model.Actor) *model.Data { return actor.AuthCredential }},
		{name: "actor info", selectData: func(actor *model.Actor) *model.Data { return actor.AuthInfo }},
	} {
		t.Run(test.name, func(t *testing.T) {
			oldDomain := newHashActorCredentialTestDomain(t, "token")
			newDomain := newHashActorCredentialTestDomain(t, "token")
			newActor := newDomain.Actors()[0]
			test.selectData(newActor).Sensitive = true

			fillHashes(t, oldDomain, newDomain)

			oldActor := oldDomain.Actors()[0]
			if test.selectData(oldActor).Hash == test.selectData(newActor).Hash {
				t.Fatalf("expected %s data hash to change", test.name)
			}
			if oldActor.Hash == newActor.Hash {
				t.Fatalf("expected actor hash to change when %s sensitive metadata changes", test.name)
			}
			if oldDomain.Hash() == newDomain.Hash() {
				t.Fatalf("expected domain hash to change when %s sensitive metadata changes", test.name)
			}
		})
	}

	for name, mutate := range map[string]func(*model.Method){
		"input":  func(method *model.Method) { method.ArgumentsSensitive = true },
		"output": func(method *model.Method) { method.ResultSensitive = true },
	} {
		t.Run(name, func(t *testing.T) {
			oldDomain := newHashTestDomain(t, "User service")
			newDomain := newHashTestDomain(t, "User service")
			mutate(newDomain.Services()[0].Methods[0])

			fillHashes(t, oldDomain, newDomain)

			if oldDomain.Services()[0].Methods[0].Hash == newDomain.Services()[0].Methods[0].Hash {
				t.Fatalf("expected method hash to change when whole-%s sensitive metadata changes", name)
			}
		})
	}

	t.Run("task input", func(t *testing.T) {
		oldDomain := newHashTaskTestDomain(t)
		newDomain := newHashTaskTestDomain(t)
		newDomain.Tasks()[0].Triggers[0].ArgumentsSensitive = true

		fillHashes(t, oldDomain, newDomain)

		if oldDomain.Tasks()[0].Triggers[0].Hash == newDomain.Tasks()[0].Triggers[0].Hash {
			t.Fatal("expected trigger hash to change when whole-input sensitive metadata changes")
		}
		if oldDomain.Tasks()[0].Hash == newDomain.Tasks()[0].Hash {
			t.Fatal("expected task hash to change when whole-input sensitive metadata changes")
		}
	})
}

func TestFillHashesIncludesApiBoundary(t *testing.T) {
	backend := newHashTestDomain(t, "Order service")
	api := newHashTestDomain(t, "Order service")
	backend.Services()[0].Pub = false
	api.Services()[0].Pub = false
	api.Services()[0].Api = true
	fillHashes(t, backend, api)
	if backend.Services()[0].Hash == api.Services()[0].Hash || backend.Hash() == api.Hash() {
		t.Fatal("API boundary must change service and domain hashes")
	}
	if backend.Services()[0].Methods[0].Hash != api.Services()[0].Methods[0].Hash {
		t.Fatal("API boundary must not change method signatures")
	}
}
