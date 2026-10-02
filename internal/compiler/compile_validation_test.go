package compiler

import (
	"path/filepath"
	"testing"
)

func TestCompileReturnsErrorWhenDomainNameMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "domain.skel")
	writeFile(t, path, "domain\n")
	_, err := Compile(Option{SkelIn: path})
	expectErrorContains(t, err, path)
}

func TestCompileReturnsErrorWhenSkelDomainMismatches(t *testing.T) {
	dir := validationInputForTest(t,
		describedUserDomain,
		"domain demo.account\ndata User { id: string }\n",
	)
	_, err := Compile(Option{SkelIn: dir})
	expectErrorContains(t, err, "domain mismatch")
}

func TestCompileDirectorySkelFileWithoutDomainReturnsError(t *testing.T) {
	dir := validationInputForTest(t,
		describedUserDomain,
		"data User { id: string }\n",
	)
	_, err := Compile(Option{SkelIn: dir})
	expectErrorContains(t, err, "missing domain declaration")
}

func TestCompileReturnsErrorWhenDirectorySkelFileDeclaresDomainDecorator(t *testing.T) {
	dir := validationInputForTest(t,
		describedUserDomain,
		"@desc(\"Not allowed\")\ndomain demo.user\ndata User { id: string }\n",
	)
	_, err := Compile(Option{SkelIn: dir})
	expectErrorContains(t, err, "domain decorator is only allowed in domain.skel")
}

func TestCompileReturnsErrorWhenDomainFileDeclaresEntries(t *testing.T) {
	dir := validationInputForTest(t,
		"@desc(\"User domain\")\ndomain demo.user\ndata User { id: string }\n",
		"domain demo.user\nservice UserService { method ping {} }\n",
	)
	_, err := Compile(Option{SkelIn: dir})
	expectErrorContains(t, err, "can only contain domain declaration and @desc")
}

func validationInputForTest(t *testing.T, domain, source string) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "domain.skel"), domain)
	writeFile(t, filepath.Join(dir, "user.skel"), source)
	return dir
}

func TestCompileSingleSkelAllowsDomainDecorator(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "user.skel")
	writeFile(t, filePath, `@desc("User domain")
domain demo.user
data User { id: string }
`)

	result, err := Compile(Option{SkelIn: filePath})
	if err != nil {
		t.Fatalf("compile file: %v", err)
	}
	domain := result.Domain
	if domain.Name() != "demo.user" {
		t.Fatalf("unexpected domain name: %s", domain.Name())
	}
	if domain.Description() != "User domain" {
		t.Fatalf("unexpected domain description: %q", domain.Description())
	}
	if len(domain.Data()) != 1 || domain.Data()[0].Name != "User" {
		t.Fatalf("unexpected data: %+v", domain.Data())
	}
}
