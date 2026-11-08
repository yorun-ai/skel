package formatter

import (
	"strings"
	"testing"
)

func TestSourceSeparatesStatements(t *testing.T) {
	cases := []struct{ name, source, want string }{
		{"fields", `data User { name: string age: int metadata: map<string, list<string?>>? }`, `data User {
    name: string
    age: int
    metadata: map<string, list<string?>>?
}`},
		{"triggers", `task JobTask { trigger first {} trigger second { input { method: string task: int } } }`, `task JobTask {
    trigger first {}

    trigger second {
        input {
            method: string
            task: int
        }
    }
}`},
		{"declarations", `actor AActor {} pub actor BActor {} @desc("中文 task {}") data User { id: uuid }`, `actor AActor {}

pub actor BActor {}

@desc("中文 task {}")
data User {
    id: uuid
}`},
		{"methods", `pub service UserService { method first {} @desc("second") method second { input { name: string age: int } output string } }`, `pub service UserService {
    method first {}

    @desc("second")
    method second {
        input {
            name: string
            age: int
        }
        output string
    }
}`},
		{"comments", "data User { name: string /* keep */ @desc(\"年龄\") age: int // trailing\n id: uuid }", `data User {
    name: string /* keep */
    @desc("年龄")
    age: int // trailing
    id: uuid
}`},
		{"enum", `enum Status { FIRST SECOND THIRD }`, `enum Status {
    FIRST
    SECOND
    THIRD
}`},
		{"resource", `resource User { action read action write { check owner { input { id: uuid tenant: string } } } }`, `resource User {
    action read
    action write {
        check owner {
            input {
                id: uuid
                tenant: string
            }
        }
    }
}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := []byte("domain demo\n" + tc.source + "\n")
			want := "domain demo\n\n" + tc.want + "\n"
			got, err := ValidatedSource(source)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != want {
				t.Fatalf("got:\n%s\nwant:\n%s", got, want)
			}
			if second := formatTestSource(t, got); string(second) != want {
				t.Fatalf("not idempotent:\n%s", second)
			}
			if parseDomainHash(t, "before.skel", source) != parseDomainHash(t, "after.skel", got) {
				t.Fatal("format changed semantic hash")
			}
			crlf := formatTestSource(t, []byte(strings.ReplaceAll(string(source), "\n", "\r\n")))
			if string(crlf) != want {
				t.Fatalf("CRLF output:\n%s", crlf)
			}
		})
	}
}

func TestSourceSeparatesInlineImports(t *testing.T) {
	source := []byte("domain demo import common.types as common import common.auth actor AActor {} actor BActor {}")
	want := "domain demo\n\nimport common.types as common\nimport common.auth\n\nactor AActor {}\n\nactor BActor {}\n"
	got, err := ValidatedSource(source)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("got:\n%s", got)
	}
	if second := formatTestSource(t, got); string(second) != want {
		t.Fatalf("not idempotent:\n%s", second)
	}
}

func TestSourceSeparatesCompleteInlineDomain(t *testing.T) {
	source := []byte(strings.ReplaceAll(string(readTestFile(t, "complete.golden.skel")), "\n", " "))
	got, err := ValidatedSource(source)
	if err != nil {
		t.Fatal(err)
	}
	if second := formatTestSource(t, got); string(second) != string(got) {
		t.Fatalf("not idempotent:\n%s\nsecond:\n%s", got, second)
	}
	checkTestSource(t, "before.skel", source)
	checkTestSource(t, "after.skel", got)
	if !strings.Contains(string(got), "require any(User:read, User:update:owner(actor.info.subject))") {
		t.Fatalf("permission expression changed:\n%s", got)
	}
}
