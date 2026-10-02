// Package loader discovers Skel inputs and reads immutable source revisions
// through filesystem, memory, or pinned Git providers. It does not parse syntax.
package loader

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"go.yorun.ai/skelc/internal/source"
)

const DomainFileName = "domain.skel"

const (
	WarningCodeDirectory   = "loader.ignored-directory"
	WarningCodeHiddenFile  = "loader.ignored-hidden-file"
	WarningCodeUnsupported = "loader.ignored-file"
)

type Warning struct {
	Code    string
	Path    string
	Message string
}

func (w Warning) Error() string { return w.Message }

type SourceFile struct {
	Document *source.Document
	FilePath string
	Content  []byte
}

type Result struct {
	Files    []*SourceFile
	Warnings []Warning
	IsDir    bool
}

type _Loader struct {
	skelIn   string
	ctx      context.Context
	provider Provider

	requireDomainFile bool
	directory         bool
	domainFile        string
	skelFiles         []string
	warnings          []Warning
}

func Load(skelIn string) (Result, error) {
	return LoadFrom(context.Background(), FileSystem{}, skelIn)
}

// LoadFrom discovers the same input contract across disk and immutable providers.
func LoadFrom(ctx context.Context, provider Provider, skelIn string) (Result, error) {
	return loadFrom(ctx, provider, skelIn, true)
}

// LoadWorkspaceFrom also accepts a directory of independent single-file inputs.
func LoadWorkspaceFrom(ctx context.Context, provider Provider, skelIn string) (Result, error) {
	return loadFrom(ctx, provider, skelIn, false)
}

func loadFrom(ctx context.Context, provider Provider, skelIn string, requireDomain bool) (Result, error) {
	sourceLoader := new(_Loader{skelIn: skelIn, ctx: ctx, provider: provider, requireDomainFile: requireDomain})
	if err := sourceLoader.discoverFiles(); err != nil {
		return Result{}, err
	}
	files, err := sourceLoader.loadSourceFiles()
	if err != nil {
		return Result{}, err
	}
	return Result{
		Files:    files,
		Warnings: append([]Warning{}, sourceLoader.warnings...),
		IsDir:    sourceLoader.directory,
	}, nil
}

func classifyFile(fileName string) string {
	switch {
	case strings.HasPrefix(fileName, "."):
		return "hidden"
	case fileName == DomainFileName:
		return "domain"
	case strings.HasSuffix(fileName, ".skel"):
		return "skel"
	default:
		return "other"
	}
}

func (l *_Loader) discoverFiles() error {
	skelInPath, err := filepath.Abs(l.skelIn)
	if err != nil {
		return fmt.Errorf("resolve skel input: %w", err)
	}
	info, err := l.provider.Stat(l.ctx, skelInPath)
	if err != nil {
		return fmt.Errorf("stat skel input %s: %w", skelInPath, err)
	}

	l.directory = info.Directory
	if !l.directory {
		return l.discoverSingleFile(skelInPath)
	}
	return l.discoverDirectory(skelInPath)
}

func (l *_Loader) discoverSingleFile(filePath string) error {
	if !strings.HasSuffix(filePath, ".skel") {
		return fmt.Errorf("%s is not a .skel file", filePath)
	}
	l.domainFile = filePath
	l.skelFiles = []string{filePath}
	return nil
}

func (l *_Loader) discoverDirectory(dirPath string) error {
	entries, err := l.provider.ReadDir(l.ctx, dirPath)
	if err != nil {
		return fmt.Errorf("read directory %s: %w", dirPath, err)
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name)
		fileKind := classifyFile(entry.Name)
		switch {
		case entry.Directory:
			l.warn(WarningCodeDirectory, fullPath, fmt.Sprintf("%s ignored (DIRECTORY)", fullPath))
		case fileKind == "hidden":
			l.warn(WarningCodeHiddenFile, fullPath, fmt.Sprintf("%s ignored (HIDDEN_FILE)", fullPath))
		case fileKind == "domain":
			l.domainFile = fullPath
			l.skelFiles = append(l.skelFiles, fullPath)
		case fileKind == "skel":
			l.skelFiles = append(l.skelFiles, fullPath)
		default:
			l.warn(WarningCodeUnsupported, fullPath, fmt.Sprintf("%s ignored", fullPath))
		}
	}
	sort.Strings(l.skelFiles)
	if l.domainFile == "" && l.requireDomainFile {
		return fmt.Errorf("%s not found under %s", DomainFileName, dirPath)
	}
	return nil
}

func (l *_Loader) warn(code, path, message string) {
	l.warnings = append(l.warnings, Warning{Code: code, Path: path, Message: message})
}

func (l *_Loader) loadSourceFiles() ([]*SourceFile, error) {
	files := make([]*SourceFile, 0, len(l.skelFiles))
	for _, skelFile := range l.skelFiles {
		document, err := l.provider.Read(l.ctx, skelFile)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", skelFile, err)
		}
		files = append(files, &SourceFile{FilePath: skelFile, Content: []byte(document.Text()), Document: document})
	}
	return files, nil
}
