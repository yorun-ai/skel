package schema

import "testing"

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
			makeDocument := func(added bool) *Document {
				data := dataDeclaration("Value", "id")
				data.Pub = test.public
				if added {
					data.Data.Members = append(data.Data.Members, &Member{Name: "extra", Type: scalarType("string")})
				}
				wrapper := dataDeclaration("Envelope")
				wrapper.Pub = false
				wrapper.Data.Members = []*Member{{Name: "value", Type: &Type{Kind: TypeKindData, Name: data.SkelName}}, {Name: "next", Type: &Type{Kind: TypeKindData, Name: wrapper.SkelName, Nullable: true}}}
				service := serviceDeclaration("Service", "read")
				ref := &Type{Kind: TypeKindData, Name: wrapper.SkelName}
				if test.input {
					service.Service.Methods[0].Arguments = []*Argument{{Name: "value", Type: ref}}
				}
				if test.output {
					service.Service.Methods[0].Result = ref
				}
				return newTestDocument(data, wrapper, service)
			}
			report, err := Diff(makeDocument(false), makeDocument(true))
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
			makeDocument := func(nullable bool) *Document {
				service := serviceDeclaration("Service", "read")
				kind := scalarType("string")
				kind.Nullable = nullable
				if input {
					service.Service.Methods[0].Arguments = []*Argument{{Name: "value", Type: kind}}
				} else {
					service.Service.Methods[0].Result = kind
				}
				return newTestDocument(service)
			}
			report, err := Diff(makeDocument(!widening), makeDocument(widening))
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
