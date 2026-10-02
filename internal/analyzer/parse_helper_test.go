package analyzer

import (
	"testing"

	"go.yorun.ai/skelc/internal/model"
	"go.yorun.ai/skelc/internal/parser/grammar"
)

func requireNoReporterDiagnostics(t *testing.T, reporter *_DiagnosticReporter) {
	t.Helper()
	if diagnostics := reporter.result(); len(diagnostics) > 0 {
		t.Fatalf("unexpected diagnostics: %v", diagnostics)
	}
}

func parseActorTest(t *testing.T, value *grammar.Actor) *model.Actor {
	t.Helper()
	reporter := newDiagnosticReporter()
	parsed, _ := parseActor(reporter, value)
	requireNoReporterDiagnostics(t, reporter)
	return parsed
}

func expectActorDiagnostic(t *testing.T, expected string, value *grammar.Actor) {
	t.Helper()
	reporter := newDiagnosticReporter()
	parseActor(reporter, value)
	assertDiagnosticsContain(t, reporter.result(), expected)
}

func parseEnumTest(t *testing.T, value *grammar.Enum) *model.Enum {
	t.Helper()
	reporter := newDiagnosticReporter()
	parsed, _ := parseEnum(reporter, value)
	requireNoReporterDiagnostics(t, reporter)
	return parsed
}

func expectEnumDiagnostic(t *testing.T, expected string, value *grammar.Enum) {
	t.Helper()
	reporter := newDiagnosticReporter()
	parseEnum(reporter, value)
	assertDiagnosticsContain(t, reporter.result(), expected)
}

func parseDataTest(t *testing.T, value *grammar.Data) *model.Data {
	t.Helper()
	reporter := newDiagnosticReporter()
	parsed, _ := parseData(reporter, value)
	requireNoReporterDiagnostics(t, reporter)
	return parsed
}

func expectDataDiagnostic(t *testing.T, expected string, value *grammar.Data) {
	t.Helper()
	reporter := newDiagnosticReporter()
	parseData(reporter, value)
	assertDiagnosticsContain(t, reporter.result(), expected)
}

func parseConfigTest(t *testing.T, value *grammar.Data) *model.Data {
	t.Helper()
	reporter := newDiagnosticReporter()
	parsed, _ := parseConfig(reporter, value)
	requireNoReporterDiagnostics(t, reporter)
	return parsed
}

func expectConfigDiagnostic(t *testing.T, expected string, value *grammar.Data) {
	t.Helper()
	reporter := newDiagnosticReporter()
	parseConfig(reporter, value)
	assertDiagnosticsContain(t, reporter.result(), expected)
}

func parseEventTest(t *testing.T, value *grammar.Event) *model.Data {
	t.Helper()
	reporter := newDiagnosticReporter()
	parsed, _ := parseEvent(reporter, value)
	requireNoReporterDiagnostics(t, reporter)
	return parsed
}

func expectEventDiagnostic(t *testing.T, expected string, value *grammar.Event) {
	t.Helper()
	reporter := newDiagnosticReporter()
	parseEvent(reporter, value)
	assertDiagnosticsContain(t, reporter.result(), expected)
}

func parseServiceTest(t *testing.T, value *grammar.Service) *model.Service {
	t.Helper()
	reporter := newDiagnosticReporter()
	parsed, _ := parseService(reporter, value)
	requireNoReporterDiagnostics(t, reporter)
	return parsed
}

func expectServiceDiagnostic(t *testing.T, expected string, value *grammar.Service) {
	t.Helper()
	reporter := newDiagnosticReporter()
	parseService(reporter, value)
	assertDiagnosticsContain(t, reporter.result(), expected)
}

func parseWebTest(t *testing.T, value *grammar.Web, pub bool) *model.Web {
	t.Helper()
	reporter := newDiagnosticReporter()
	parsed, _ := parseWeb(reporter, value, pub)
	requireNoReporterDiagnostics(t, reporter)
	return parsed
}

func expectWebDiagnostic(t *testing.T, expected string, value *grammar.Web, pub bool) {
	t.Helper()
	reporter := newDiagnosticReporter()
	parseWeb(reporter, value, pub)
	assertDiagnosticsContain(t, reporter.result(), expected)
}

func parseTaskTest(t *testing.T, value *grammar.Task) *model.Task {
	t.Helper()
	reporter := newDiagnosticReporter()
	parsed, _ := parseTask(reporter, value)
	requireNoReporterDiagnostics(t, reporter)
	return parsed
}

func expectTaskDiagnostic(t *testing.T, expected string, value *grammar.Task) {
	t.Helper()
	reporter := newDiagnosticReporter()
	parseTask(reporter, value)
	assertDiagnosticsContain(t, reporter.result(), expected)
}

func parseTypeTest(t *testing.T, value *grammar.Type) *model.Type {
	t.Helper()
	reporter := newDiagnosticReporter()
	parsed, _ := parseType(reporter, value)
	requireNoReporterDiagnostics(t, reporter)
	return parsed
}

func fixTypeRefTest(t *testing.T, value *model.Type, refs *_RefContext) {
	t.Helper()
	reporter := newDiagnosticReporter()
	fixTypeRef(reporter, value, refs)
	requireNoReporterDiagnostics(t, reporter)
}

func expectFixTypeRefDiagnostic(t *testing.T, expected string, value *model.Type, refs *_RefContext) {
	t.Helper()
	reporter := newDiagnosticReporter()
	fixTypeRef(reporter, value, refs)
	assertDiagnosticsContain(t, reporter.result(), expected)
}
