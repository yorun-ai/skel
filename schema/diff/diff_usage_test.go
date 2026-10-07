package diff

import (
	"testing"

	"go.yorun.ai/skel/schema"
)

func TestDiffDataUsageDirections(t *testing.T) {
	for _, test := range []struct {
		name                  string
		input, output, public bool
		impact                ImpactLevel
	}{
		{"output", false, true, false, ImpactCompatible},
		{"input", true, false, false, ImpactBreaking},
		{"both", true, true, false, ImpactBreaking},
		{"independent public use", false, true, true, ImpactBreaking},
		{"unknown", false, false, false, ImpactBreaking},
	} {
		t.Run(test.name, func(t *testing.T) {
			makeDomain := func(added bool) *schema.Domain {
				data := dataDeclaration("Value", "id")
				data.Pub = test.public
				if added {
					data.Data.Members = append(data.Data.Members, &schema.DataMember{Name: "extra", Type: scalarType("string")})
				}
				wrapper := dataDeclaration("Envelope")
				wrapper.Pub = false
				wrapper.Data.Members = []*schema.DataMember{{Name: "value", Type: &schema.Type{Kind: schema.TypeKindData, SkelName: data.SkelName}}, {Name: "next", Type: &schema.Type{Kind: schema.TypeKindData, SkelName: wrapper.SkelName, Nullable: true}}}
				service := serviceDeclaration("Service", "read")
				ref := &schema.Type{Kind: schema.TypeKindData, SkelName: wrapper.SkelName}
				if test.input {
					service.Service.Methods[0].Arguments = []*schema.Argument{{Name: "value", Type: ref}}
				}
				if test.output {
					service.Service.Methods[0].ResultType = ref
				}
				return newTestDomain(data, wrapper, service)
			}
			report, err := Compare(makeDomain(false), makeDomain(true))
			if err != nil {
				t.Fatal(err)
			}
			if len(report.Changes) != 1 || report.Changes[0].Impact != test.impact {
				t.Fatalf("unexpected report: %+v", report)
			}
		})
	}
}

func TestDiffNullableByRole(t *testing.T) {
	for _, input := range []bool{false, true} {
		for _, widening := range []bool{false, true} {
			makeDomain := func(nullable bool) *schema.Domain {
				service := serviceDeclaration("Service", "read")
				kind := scalarType("string")
				kind.Nullable = nullable
				if input {
					service.Service.Methods[0].Arguments = []*schema.Argument{{Name: "value", Type: kind}}
				} else {
					service.Service.Methods[0].ResultType = kind
				}
				return newTestDomain(service)
			}
			report, err := Compare(makeDomain(!widening), makeDomain(widening))
			if err != nil {
				t.Fatal(err)
			}
			want := ImpactBreaking
			if input == widening {
				want = ImpactCompatible
			}
			if len(report.Changes) != 1 || report.Changes[0].Impact != want {
				t.Fatalf("input=%v widening=%v: %+v", input, widening, report)
			}
		}
	}
}
