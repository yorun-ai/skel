package sourcediff

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	compiler "go.yorun.ai/skel/internal/compiler"
	"go.yorun.ai/skel/internal/loader"
	"go.yorun.ai/skel/schema"
)

// ErrGitHistoryUnavailable identifies a domain for which no usable Git HEAD
// baseline exists. Continuous editor diagnostics may ignore this condition.
var ErrGitHistoryUnavailable = errors.New("git history unavailable")

func compileGitBaseline(ctx context.Context, differ *Differ, candidate compiler.WorkspaceDomain) (*schema.Domain, string, error) {
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
	requireDomainFile := false
	for _, input := range candidate.Sources {
		requireDomainFile = requireDomainFile || input.DirectoryInput
	}
	cacheKey := repositoryRoot + "\x00" + filepath.Clean(relativeRoot) + "\x00" + candidate.Name + fmt.Sprintf("\x00%t", requireDomainFile)
	if domain, cachedRoot, cachedErr, ok := differ.cachedBaseline(cacheKey, head); ok {
		return domain, cachedRoot, cachedErr
	}
	provider, err := loader.NewGit(ctx, repositoryRoot, head, root)
	if err != nil {
		if ctx.Err() != nil {
			return nil, "", ctx.Err()
		}
		return nil, "", gitHistoryError(root, err)
	}
	if _, err = provider.Stat(ctx, root); err != nil {
		if ctx.Err() != nil {
			return nil, "", ctx.Err()
		}
		failure := gitHistoryError(root, err)
		differ.storeBaseline(cacheKey, head, repositoryRoot, nil, failure)
		return nil, repositoryRoot, failure
	}
	diagnostics, domains, err := compiler.AnalyzeInputFromContext(ctx, provider, root, requireDomainFile)
	var domain *schema.Domain
	for _, analyzed := range domains {
		if analyzed.Name != candidate.Name {
			continue
		}
		domain = analyzed.Schema
		break
	}
	if domain == nil && (err != nil || diagnostics.HasErrors()) {
		if ctx.Err() != nil {
			return nil, "", ctx.Err()
		}
		if err == nil {
			for i := range diagnostics {
				remapDiagnosticBaseline(&diagnostics[i], repositoryRoot)
			}
			err = diagnostics
		}
		failure := fmt.Errorf("%w: compile Git HEAD schema compatibility baseline for %s: %w", ErrSourceCompilation, candidate.Name, err)
		differ.storeBaseline(cacheKey, head, repositoryRoot, nil, failure)
		return nil, repositoryRoot, failure
	}
	if domain == nil {
		failure := gitHistoryError(root, nil)
		differ.storeBaseline(cacheKey, head, repositoryRoot, nil, failure)
		return nil, repositoryRoot, failure
	}

	if ctx.Err() != nil {
		return nil, "", ctx.Err()
	}
	differ.storeBaseline(cacheKey, head, repositoryRoot, domain, nil)
	return domain, repositoryRoot, nil
}

func gitOutput(ctx context.Context, directory string, args ...string) (string, error) {
	content, err := loader.GitBytes(ctx, directory, args...)
	return string(content), err
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
