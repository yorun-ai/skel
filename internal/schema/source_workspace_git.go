package schema

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"go.yorun.ai/skelc/internal/compiler"
)

// ErrGitHistoryUnavailable identifies a domain for which no usable Git HEAD
// baseline exists. Continuous editor diagnostics may ignore this condition.
var ErrGitHistoryUnavailable = errors.New("git history unavailable")

func projectGitBaseline(ctx context.Context, differ *SourceDiffer, candidate compiler.WorkspaceDomain) (*Document, string, error) {
	root, err := filepath.Abs(candidate.Root)
	if err != nil {
		return nil, "", gitHistoryError(candidate.Root, err)
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return nil, "", gitHistoryError(candidate.Root, err)
	}
	if err := ctx.Err(); err != nil {
		return nil, "", err
	}
	if cached := differ.cachedGitFailure(root); cached != nil {
		return nil, "", cached
	}
	directory := root
	if workspaceDomainIsFile(candidate) {
		directory = filepath.Dir(root)
	}
	repositoryIdentity, err := gitOutput(ctx, directory, "rev-parse", "--show-toplevel", "HEAD")
	if err != nil {
		if ctx.Err() != nil {
			return nil, "", ctx.Err()
		}
		failure := gitHistoryError(root, err)
		differ.storeGitFailure(root, failure)
		return nil, "", failure
	}
	differ.clearGitFailure(root)
	identityParts := strings.Split(strings.TrimSpace(repositoryIdentity), "\n")
	if len(identityParts) != 2 {
		return nil, "", gitHistoryError(root, nil)
	}
	repositoryRoot := identityParts[0]
	head := identityParts[1]
	repositoryRoot, err = filepath.EvalSymlinks(repositoryRoot)
	if err != nil {
		return nil, "", gitHistoryError(root, err)
	}
	relativeRoot, err := filepath.Rel(repositoryRoot, root)
	if err != nil || pathEscapesRoot(relativeRoot) {
		return nil, "", gitHistoryError(root, err)
	}
	cacheKey := repositoryRoot + "\x00" + filepath.Clean(relativeRoot) + "\x00" + candidate.Name
	if document, cachedRoot, cachedErr, ok := differ.cachedBaseline(cacheKey, head); ok {
		return document, cachedRoot, cachedErr
	}
	gitRoot := filepath.ToSlash(relativeRoot)
	listArgs := []string{"ls-tree", "-r", "-z", "--name-only", head}
	if relativeRoot != "." {
		listArgs = append(listArgs, "--", gitRoot)
	}
	names, err := gitOutput(ctx, repositoryRoot, listArgs...)
	if err != nil {
		return nil, "", gitHistoryError(root, err)
	}
	sources := []compiler.Source{}
	for _, name := range strings.Split(names, "\x00") {
		if name == "" || filepath.Ext(name) != ".skel" {
			continue
		}
		if workspaceDomainIsFile(candidate) {
			if filepath.Clean(filepath.FromSlash(name)) != filepath.Clean(relativeRoot) {
				continue
			}
		} else if filepath.Clean(filepath.Dir(filepath.FromSlash(name))) != filepath.Clean(relativeRoot) {
			continue
		}
		content, showErr := gitBytes(ctx, repositoryRoot, "show", head+":"+name)
		if showErr != nil {
			return nil, "", gitHistoryError(root, showErr)
		}
		sources = append(sources, compiler.Source{
			Path: filepath.Join(repositoryRoot, filepath.FromSlash(name)), Root: root, Content: content,
		})
	}
	if len(sources) == 0 {
		failure := gitHistoryError(root, nil)
		differ.storeBaseline(cacheKey, head, repositoryRoot, nil, failure)
		return nil, repositoryRoot, failure
	}
	slices.SortFunc(sources, func(left, right compiler.Source) int { return strings.Compare(left.Path, right.Path) })
	analyzer := compiler.NewWorkspaceAnalyzer()
	diagnostics, domains, err := analyzer.AnalyzeDomainsContext(ctx, sources)
	if err != nil {
		return nil, "", err
	}
	for _, domain := range domains {
		if domain.Name != candidate.Name {
			continue
		}
		document, projectErr := Project(domain.Model, nil)
		if projectErr != nil {
			return nil, "", projectErr
		}
		differ.storeBaseline(cacheKey, head, repositoryRoot, document, nil)
		return document, repositoryRoot, nil
	}
	if len(diagnostics) > 0 {
		failure := fmt.Errorf("compile Git HEAD schema compatibility baseline for %s: %s", candidate.Name, diagnostics[0].Message)
		differ.storeBaseline(cacheKey, head, repositoryRoot, nil, failure)
		return nil, repositoryRoot, failure
	}
	failure := gitHistoryError(root, nil)
	differ.storeBaseline(cacheKey, head, repositoryRoot, nil, failure)
	return nil, repositoryRoot, failure
}

func gitOutput(ctx context.Context, directory string, args ...string) (string, error) {
	content, err := gitBytes(ctx, directory, args...)
	return string(content), err
}

func gitBytes(ctx context.Context, directory string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, "git", append([]string{"-C", directory, "--literal-pathspecs"}, args...)...)
	content, err := command.Output()
	if err == nil {
		return content, nil
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		message := strings.TrimSpace(string(exitError.Stderr))
		if message != "" {
			return nil, fmt.Errorf("git %s: %s", strings.Join(args, " "), message)
		}
	}
	return nil, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
}

func gitHistoryError(root string, cause error) error {
	message := fmt.Sprintf("git history not found for schema domain source directory %s", root)
	if cause != nil {
		return fmt.Errorf("%w: %s: %v", ErrGitHistoryUnavailable, message, cause)
	}
	return fmt.Errorf("%w: %s", ErrGitHistoryUnavailable, message)
}

func pathEscapesRoot(path string) bool {
	return path == ".." || strings.HasPrefix(path, ".."+string(filepath.Separator)) || filepath.IsAbs(path)
}
