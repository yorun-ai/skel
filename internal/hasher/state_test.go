package hasher

import (
	"fmt"
	"testing"

	"go.yorun.ai/skel/internal/parser"
	"go.yorun.ai/skel/schema"
)

func TestEventDirectionChangesHashes(t *testing.T) {
	plainDomain, plainEvent := newHashDataKindTestDomain(schema.DataKindEvent)
	extDomain, extEvent := newHashDataKindTestDomain(schema.DataKindEvent)
	extEvent.Ext = true
	fillHashes(t, plainDomain, extDomain)
	if plainEvent.Hash == extEvent.Hash || plainDomain.Hash() == extDomain.Hash() {
		t.Fatal("event direction must change event and domain hashes")
	}
}

func TestFillHashesIncludesActorCredential(t *testing.T) {
	oldDomain := newHashActorCredentialTestDomain(t, "subject")
	newDomain := newHashActorCredentialTestDomain(t, "tenant")

	fillHashes(t, oldDomain, newDomain)

	if oldDomain.Actors()[0].Hash == newDomain.Actors()[0].Hash {
		t.Fatal("expected actor hash to change when credential changes")
	}
	if oldDomain.Hash() == newDomain.Hash() {
		t.Fatal("expected domain hash to change when actor credential changes")
	}
}

func TestActorIdentifierChangesHash(t *testing.T) {
	baseline := newHashActorCredentialTestDomain(t, "token")
	candidate := newHashActorCredentialTestDomain(t, "token")
	candidate.Actors()[0].Auth.IdentifierField = "userId"
	fillHashes(t, baseline, candidate)
	if baseline.Hash() == candidate.Hash() || baseline.Actors()[0].Hash == candidate.Actors()[0].Hash {
		t.Fatal("identifier change did not affect hashes")
	}
}

func TestWebMountChangesWebAndDomainHashes(t *testing.T) {
	hashes := map[string]bool{}
	domainHashes := map[string]bool{}
	for _, path := range []string{"", "/", "/portal", "/portal/", "/other"} {
		domain := newHashAllowViaTestDomain(t, "client")
		domain.Webs()[0].MountPath = path
		fillHashes(t, domain)
		hash := domain.Webs()[0].Hash
		if hashes[hash] || domainHashes[domain.Hash()] {
			t.Fatalf("mount %q did not affect compatibility hashes", path)
		}
		hashes[hash], domainHashes[domain.Hash()] = true, true
	}
}

func TestWebAuthChangesHashes(t *testing.T) {
	hashes := map[string]bool{}
	domains := map[string]bool{}
	for _, auth := range []schema.AuthMode{schema.AuthModeRequired, schema.AuthModeOptional, schema.AuthModeAnonymous, schema.AuthModeOff} {
		domain := newHashAllowViaTestDomain(t, "client")
		domain.Webs()[0].AuthMode = auth
		fillHashes(t, domain)
		hash := domain.Webs()[0].Hash
		if hashes[hash] || domains[domain.Hash()] {
			t.Fatalf("auth %q did not change hashes", auth)
		}
		hashes[hash], domains[domain.Hash()] = true, true
	}
}

func TestActorAuthResultHashDependencies(t *testing.T) {
	build := func(t *testing.T, clause string, infoType string) *schema.Domain {
		t.Helper()
		source := fmt.Sprintf(`domain demo
pub actor OpsActor { auth { credential { token: string } info { id: %s } } }
pub actor AdminActor { auth { credential { token: string } info { id: string } } }
pub actor UserActor { auth {
    %s
    credential { token: string }
    info { id: string }
} }
api service UserApiService { for UserActor auth required method ping { output string } }
web UserWeb { for UserActor auth required }
`, infoType, clause)
		content, err := parser.ParseSource("actor.skel", []byte(source))
		if err != nil {
			t.Fatal(err)
		}
		domain := analyzeHashTestDomain(t, content).Schema()
		fillHashes(t, domain)
		return domain
	}
	findActor := func(domain *schema.Domain, name string) *schema.Actor {
		for _, actor := range domain.Actors() {
			if actor.Name == name {
				return actor
			}
		}
		panic("missing actor test fixture: " + name)
	}
	for _, tc := range []struct {
		name     string
		before   string
		after    string
		infoType string
	}{
		{"enable", "", "impersonator OpsActor", "string"},
		{"add operator", "impersonator OpsActor", "impersonator OpsActor, AdminActor", "string"},
		{"change operator info", "impersonator OpsActor", "impersonator OpsActor", "int"},
		{"self and other operators", "impersonator UserActor, OpsActor", "impersonator UserActor, OpsActor", "int"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			baseline := build(t, tc.before, "string")
			candidate := build(t, tc.after, tc.infoType)
			before, after := findActor(baseline, "UserActor"), findActor(candidate, "UserActor")
			hashes := map[string][2]string{
				"actor":        {before.Hash, after.Hash},
				"result":       {before.Auth.Result.Hash, after.Auth.Result.Hash},
				"method":       {before.Auth.Method.Hash, after.Auth.Method.Hash},
				"auth service": {before.Auth.Service.Hash, after.Auth.Service.Hash},
				"API service":  {baseline.Services()[0].Hash, candidate.Services()[0].Hash},
				"web":          {baseline.Webs()[0].Hash, candidate.Webs()[0].Hash},
				"domain":       {baseline.Hash(), candidate.Hash()},
			}
			if before.Auth.ImpersonatorInfo != nil {
				hashes["impersonator info"] = [2]string{before.Auth.ImpersonatorInfo.Hash, after.Auth.ImpersonatorInfo.Hash}
			}
			for name, pair := range hashes {
				if pair[0] == pair[1] {
					t.Errorf("%s hash did not change: %s", name, pair[0])
				}
			}
			if findActor(baseline, "AdminActor").Hash != findActor(candidate, "AdminActor").Hash {
				t.Error("unaffected AdminActor hash changed")
			}
			actorHash, resultHash, domainHash := after.Hash, after.Auth.Result.Hash, candidate.Hash()
			fillHashes(t, candidate)
			if after.Hash != actorHash || after.Auth.Result.Hash != resultHash || candidate.Hash() != domainHash {
				t.Error("repeated hashing changed the contract")
			}
		})
	}
}
