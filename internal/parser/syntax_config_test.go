package parser

import "testing"

func TestAbstractConfigRejectsInvalidSyntax(t *testing.T) {
	for _, source := range []string{
		"pub abs config BaseAbsConfig eternal {}", "abs pub config BaseAbsConfig eternal {}",
		"abs config AppConfig = BaseAbsConfig", "abs data Item {}", "data Item = BaseAbsConfig",
		"config AppConfig eternal = BaseAbsConfig", "config AppConfig = BaseAbsConfig {}",
	} {
		if _, err := ParseSource("config.skel", []byte("domain demo\n"+source)); err == nil {
			t.Fatalf("accepted %s", source)
		}
	}
}
