package compiler

import (
	"testing"

	"github.com/alecthomas/participle/v2/lexer"
	"go.yorun.ai/skel/internal/parser/grammar"
)

func TestMergeDomainContentsUsesDomainFileDomain(t *testing.T) {
	domainFileContent := &grammar.SkelContent{
		Pos:    lexer.Position{Filename: "/workspace/domain.skel"},
		Domain: domainContentForTest("demo.user", "User domain"),
	}
	otherContent := &grammar.SkelContent{
		Pos:    lexer.Position{Filename: "/workspace/user.skel"},
		Domain: domainContentForTest("demo.user", ""),
		Entries: []*grammar.SkelEntry{
			{Data: &grammar.Data{Name: identForTest("User")}},
		},
	}

	merged := mergeDomainContents([]*grammar.SkelContent{otherContent, domainFileContent})
	if merged.Domain != domainFileContent.Domain {
		t.Fatal("expected merged domain to come from domain.skel content")
	}
	if len(merged.Entries) != 1 {
		t.Fatalf("unexpected merged entry count: %d", len(merged.Entries))
	}
}
