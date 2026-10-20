package analyzer

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.yorun.ai/skel/internal/parser"
)

func TestConfigAdoptionRejectsConcreteAndChainedSources(t *testing.T) {
	for _, source := range []string{
		"config BaseConfig eternal {}\nconfig AppConfig = BaseConfig",
		"pub config BaseConfig eternal {}\nconfig AppConfig = BaseConfig",
		"abs config BaseAbsConfig eternal {}\nconfig MidConfig = BaseAbsConfig\nconfig AppConfig = MidConfig",
		"config AppConfig = AppConfig",
		"data Base {}\nconfig AppConfig = Base",
		"config AppConfig = missing.BaseAbsConfig",
	} {
		t.Run(source, func(t *testing.T) {
			content, err := parser.ParseSource("config.skel", []byte("domain demo\n"+source))
			require.NoError(t, err)
			_, diagnostics := Analyze(content, nil)
			require.NotEmpty(t, diagnostics)
		})
	}
}

func TestConfigAdoptionPreservesImportedTypeOwnership(t *testing.T) {
	base, err := parser.ParseSource("base.skel", []byte("domain shared\ndata Options { value: string }\nabs config BaseAbsConfig eternal { options: list<Options> }"))
	require.NoError(t, err)
	imported, diagnostics := Analyze(base, nil)
	require.Empty(t, diagnostics)
	input, err := parser.ParseSource("app.skel", []byte("domain app\nimport shared as s\nconfig AppConfig = s.BaseAbsConfig"))
	require.NoError(t, err)
	analysis, diagnostics := Analyze(input, []*Analysis{imported})
	require.Empty(t, diagnostics)
	config := analysis.Schema().Configs()[0]
	require.Equal(t, "shared.BaseAbsConfig", config.FromSkelName())
	require.Equal(t, "shared", config.Members[0].Type.List.Element.ExternalDomain)
	require.Equal(t, "s", config.Members[0].Type.List.Element.ExternalAlias)
	require.Empty(t, imported.Schema().Configs()[0].Members[0].Type.List.Element.ExternalDomain)
}

func TestConfigAdoptionRejectsAbstractSuffixOnConcreteName(t *testing.T) {
	for _, prefix := range []string{"", "pub "} {
		for _, name := range []string{"PrimaryAbsConfig", "PrimaryAbsConfigV2"} {
			content, err := parser.ParseSource("config.skel", []byte("domain demo\nabs config BaseAbsConfig eternal {}\n"+prefix+"config "+name+" = BaseAbsConfig"))
			require.NoError(t, err)
			_, diagnostics := Analyze(content, nil)
			require.NotEmpty(t, diagnostics)
			require.Contains(t, diagnostics[0].Error(), "AbsConfig suffix is reserved for abs config")
		}
	}
}
