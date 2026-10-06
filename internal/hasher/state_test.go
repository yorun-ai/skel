package hasher

import (
	"go.yorun.ai/skel/internal/model"
	"testing"
)

func TestEventDirectionChangesHashes(t *testing.T) {
	plainDomain, plainEvent := newHashDataKindTestDomain(model.DataKindEvent)
	extDomain, extEvent := newHashDataKindTestDomain(model.DataKindEvent)
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
	candidate.Actors()[0].IdentifierField = "userId"
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
	for _, auth := range []model.AuthMode{model.AuthModeUnset, model.AuthModeAuth, model.AuthModeNoAuth, model.AuthModeRequired, model.AuthModeOptional, model.AuthModeAnonymous, model.AuthModeOff} {
		domain := newHashAllowViaTestDomain(t, "client")
		domain.Webs()[0].Auth = auth
		fillHashes(t, domain)
		hash := domain.Webs()[0].Hash
		if hashes[hash] || domains[domain.Hash()] {
			t.Fatalf("auth %q did not change hashes", auth)
		}
		hashes[hash], domains[domain.Hash()] = true, true
	}
}
