// Package index builds immutable language indexes for Skel documents.
package index

import (
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.yorun.ai/skelc/internal/compiler"
	"go.yorun.ai/skelc/internal/lsp/source"
	"go.yorun.ai/skelc/internal/parser/grammar"
)

type Document struct {
	URI              uri.URI
	Path             string
	Source           string
	Buffer           source.Buffer
	Version          int32
	Open             bool
	Domain           string
	Imports          map[string]string
	Definitions      []Definition
	Symbols          []Symbol
	Occurrences      []Occurrence
	ParseDiagnostics compiler.Diagnostics
	Parsed           *grammar.SkelContent
}

type Definition struct {
	// Confirmed means this declaration is present in successfully recovered syntax.
	Confirmed   bool
	Key         string
	Name        string
	Detail      string
	Description string
	Deprecated  bool
	Kind        protocol.SymbolKind
	Range       protocol.Range
}

type Occurrence struct {
	Key   string
	Range protocol.Range
}

type Symbol struct {
	Name        string
	Detail      string
	Description string
	Deprecated  bool
	Kind        protocol.SymbolKind
	Range       protocol.Range
	Children    []Symbol
}

// Build parses and indexes one in-memory Skel document.
func Build(documentURI uri.URI, path, content string, version int32) *Document {
	if path == "" {
		path = documentURI.FsPath()
	}
	document := &Document{URI: documentURI, Path: path, Source: content, Version: version, Buffer: source.New(content), Imports: map[string]string{}}
	parsed, diagnostics := compiler.ParseSourceRecovering(document.AnalysisPath(), []byte(content))
	document.Parsed = parsed
	document.ParseDiagnostics = diagnostics
	if len(diagnostics) > 0 {
		indexIncompleteDocument(document, document.Buffer.IdentifierTokens())
		recovered := map[protocol.Range]bool{}
		if parsed != nil {
			for _, entry := range parsed.Entries {
				name, pos, _, _ := entryDefinition(entry)
				if name != "" {
					recovered[identifierRange(document.Buffer, pos, name)] = true
				}
			}
		}
		for i := range document.Definitions {
			document.Definitions[i].Confirmed = recovered[document.Definitions[i].Range]
		}
		document.Occurrences = indexOccurrences(document)
		return document
	}
	if parsed.Domain != nil && parsed.Domain.Name != nil {
		document.Domain = parsed.Domain.Name.String()
	}
	for _, importDecl := range parsed.Imports {
		domain := importDecl.Domain.String()
		alias := domain
		if importDecl.Alias != nil {
			alias = importDecl.Alias.Value
		}
		document.Imports[alias] = domain
	}
	for _, entry := range parsed.Entries {
		name, pos, kind, detail := entryDefinition(entry)
		if name == "" {
			continue
		}
		range_ := identifierRange(document.Buffer, pos, name)
		description, deprecated := documentationFromDecoratorGroups(entry.Decorators)
		document.Definitions = append(document.Definitions, Definition{
			Confirmed: true, Key: document.Domain + "." + name, Name: name, Detail: detail, Description: description,
			Deprecated: deprecated, Kind: kind, Range: range_,
		})
		document.Symbols = append(document.Symbols, entrySymbol(document.Buffer, entry, name, detail, description, deprecated, kind, range_))
	}
	document.Occurrences = indexOccurrences(document)
	return document
}

// AnalysisPath is a unique source identity; Path remains the local filesystem path.
func (d *Document) AnalysisPath() string {
	if !d.URI.IsFile() {
		return string(d.URI)
	}
	return d.Path
}
