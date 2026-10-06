package skelc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	ucli "github.com/urfave/cli/v3"
	skelapi "go.yorun.ai/skel/api"
	"go.yorun.ai/skel/internal/cmd/skelc/output"
	"go.yorun.ai/skel/internal/codegen"
	internalcompiler "go.yorun.ai/skel/internal/compiler"
	schemas "go.yorun.ai/skel/schema"
)

const (
	commandSchema         = "schema"
	commandSchemaList     = "list"
	commandSchemaGet      = "get"
	commandSchemaSnapshot = "snapshot"
	commandSchemaDiff     = "diff"

	flagSchemaSkelIn         = "skel-in"
	flagSchemaBaselineSkelIn = "baseline-skel-in"
)

func newSchemaCommand() *ucli.Command {
	return &ucli.Command{
		Name:               commandSchema,
		Usage:              "inspect, snapshot, and diff skel schemas",
		HideHelpCommand:    true,
		CustomHelpTemplate: groupCommandHelpTemplate,
		Commands: []*ucli.Command{
			newSchemaListCommand(),
			newSchemaDepCommand(),
			newSchemaGetCommand(),
			newSchemaSnapshotCommand(),
			newSchemaDiffCommand(),
		},
	}
}

func newSchemaListCommand() *ucli.Command {
	return &ucli.Command{
		Name:      commandSchemaList,
		Usage:     "list top-level schema declarations",
		ArgsUsage: "[TYPE]",
		Flags:     newSchemaListFlags(),
		Action: func(_ context.Context, cmd *ucli.Command) error {
			kind, err := parseSchemaListKind(cmd)
			if err != nil {
				return commandFailure(output.ErrorCodeInvalidArgument, err)
			}
			document, err := loadSchemaList(cmd)
			if err != nil {
				return err
			}
			entries := filterSchemaEntries(schemas.Entries(document), kind)
			return writeSchemaResult(cmd, entries, "schema declarations")
		},
	}
}

func newSchemaGetCommand() *ucli.Command {
	return &ucli.Command{
		Name:      commandSchemaGet,
		Usage:     "get one complete top-level schema declaration",
		ArgsUsage: "TYPE SKEL_NAME",
		Flags:     newSchemaGetFlags(),
		Action: func(_ context.Context, cmd *ucli.Command) error {
			kind, skelName, err := parseSchemaGetArguments(cmd)
			if err != nil {
				return commandFailure(output.ErrorCodeInvalidArgument, err)
			}
			document, err := loadQuerySchema(cmd)
			if err != nil {
				return err
			}
			declaration := schemas.Find(document, schemas.DeclarationType(kind), skelName)
			return writeSchemaResult(cmd, declaration, "schema declaration")
		},
	}
}

func newSchemaSnapshotCommand() *ucli.Command {
	return &ucli.Command{
		Name:  commandSchemaSnapshot,
		Usage: "output a normalized schema snapshot",
		Flags: []ucli.Flag{
			&ucli.StringFlag{Name: flagSchemaSkelIn, Usage: "skeleton input file or directory"},
		},
		Action: func(_ context.Context, cmd *ucli.Command) error {
			if cmd.Args().Len() != 0 {
				return commandFailure(output.ErrorCodeInvalidArgument,
					fmt.Errorf("unexpected args for %s %s", commandSchema, commandSchemaSnapshot))
			}
			document, err := loadSourceSchema(cmd, flagSchemaSkelIn, cmd.String(flagSchemaSkelIn))
			if err != nil {
				return err
			}
			if err := schemas.Validate(document); err != nil {
				return commandFailure(output.ErrorCodeCommandFailed, err)
			}
			return writeSchemaResult(cmd, document, "schema snapshot")
		},
	}
}

func newSchemaDiffCommand() *ucli.Command {
	return &ucli.Command{
		Name:  commandSchemaDiff,
		Usage: "list all schema changes between baseline and candidate source",
		Flags: []ucli.Flag{
			&ucli.StringFlag{Name: flagSchemaSkelIn, Usage: "candidate skeleton input file or directory"},
			&ucli.StringFlag{Name: flagSchemaBaselineSkelIn, Usage: "baseline skeleton input file or directory; defaults to the candidate path at Git HEAD"},
		},
		Action: func(ctx context.Context, cmd *ucli.Command) error {
			if cmd.Args().Len() != 0 {
				return commandFailure(output.ErrorCodeInvalidArgument,
					fmt.Errorf("unexpected args for %s %s", commandSchema, commandSchemaDiff))
			}
			candidateOption := internalcompiler.Option{SkelIn: cmd.String(flagSchemaSkelIn), Strict: cmd.Bool(flagStrict)}
			if err := normalizeCompilerOption(&candidateOption); err != nil {
				return commandFailure(output.ErrorCodeInvalidArgument, err)
			}
			baselineSkelIn := strings.TrimSpace(cmd.String(flagSchemaBaselineSkelIn))
			if baselineSkelIn != "" {
				baselineOption := internalcompiler.Option{SkelIn: baselineSkelIn}
				if err := normalizeCompilerOption(&baselineOption); err != nil {
					return commandFailure(output.ErrorCodeInvalidArgument, err)
				}
				baselineSkelIn = baselineOption.SkelIn
			}
			option := skelapi.SchemaDiffOption{}
			if baselineSkelIn != "" {
				option.Baseline = &skelapi.Input{SkelIn: baselineSkelIn}
			}
			report, err := skelapi.DiffSchemaSourcesContext(ctx, skelapi.Input{SkelIn: candidateOption.SkelIn, Strict: cmd.Bool(flagStrict)}, option)
			if err != nil {
				switch {
				case errors.Is(err, skelapi.ErrGitHistoryUnavailable):
					return commandFailure(output.ErrorCodeGitHistoryNotFound,
						fmt.Errorf("%w; pass an explicit --%s", err, flagSchemaBaselineSkelIn))
				case errors.Is(err, skelapi.ErrSchemaSourceCompilation):
					return commandFailure(output.ErrorCodeCompilationFailed, err)
				default:
					return commandFailure(output.ErrorCodeCommandFailed, err)
				}
			}
			return writeSchemaResult(cmd, report, "schema diff")
		},
	}
}

func newSchemaListFlags() []ucli.Flag {
	return []ucli.Flag{
		&ucli.StringFlag{Name: flagSchemaSkelIn, Usage: "skeleton input file or directory"},
		&ucli.BoolFlag{Name: flagGenPub, Usage: "list backend public contract declarations and their local type dependencies"},
		&ucli.BoolFlag{Name: flagGenApi, Usage: "list API generation declarations"},
		&ucli.BoolFlag{Name: flagGenPrune, Usage: "retain only selected API roots and their type dependencies; requires --api"},
		&ucli.StringSliceFlag{Name: flagGenActor, Usage: "fully qualified API actor name; requires --api, repeat to select multiple actors"},
		&ucli.StringSliceFlag{Name: flagGenName, Usage: "fully qualified local data or enum root; requires --api --prune, repeat to select multiple types"},
		&ucli.StringSliceFlag{Name: flagGenSkelImport, Usage: "skel dependency mapping in domain=path form; requires --pub or --api, repeat for transitive imports"},
	}
}

func loadSchemaList(cmd *ucli.Command) (*schemas.Document, error) {
	api, pub := cmd.Bool(flagGenApi), cmd.Bool(flagGenPub)
	if api && pub {
		return nil, commandFailure(output.ErrorCodeInvalidArgument, fmt.Errorf("flags api and pub are mutually exclusive"))
	}
	selection := codegen.ApiFilter{Actors: cmd.StringSlice(flagGenActor), Prune: cmd.Bool(flagGenPrune), Types: cmd.StringSlice(flagGenName)}
	if err := codegen.ValidateApiFilterMode(selection, api); err != nil {
		return nil, generationCommandFailure(err)
	}
	if !api && !pub {
		if cmd.IsSet(flagGenSkelImport) {
			return nil, commandFailure(output.ErrorCodeInvalidArgument, fmt.Errorf("flag skel-import requires api or pub"))
		}
		return loadQuerySchema(cmd)
	}
	imports, err := parseMappingFlags(cmd.StringSlice(flagGenSkelImport), flagGenSkelImport)
	if err != nil {
		return nil, commandFailure(output.ErrorCodeInvalidArgument, err)
	}
	result, err := skelapi.QuerySchema(skelapi.Input{SkelIn: cmd.String(flagSchemaSkelIn), SkelImports: imports, Strict: cmd.Bool(flagStrict)}, skelapi.SchemaQueryOption{Api: api, Pub: pub, ApiFilter: selection})
	if err != nil {
		return nil, generationCommandFailure(err)
	}
	writeWarningLogs(cmd, result.Diagnostics)
	return result.Document, nil
}

func newSchemaGetFlags() []ucli.Flag {
	return []ucli.Flag{
		&ucli.StringFlag{Name: flagSchemaSkelIn, Usage: "skeleton input file or directory"},
	}
}

func parseSchemaListKind(cmd *ucli.Command) (string, error) {
	if cmd.Args().Len() > 1 {
		return "", fmt.Errorf("unexpected args for %s %s", commandSchema, commandSchemaList)
	}
	if cmd.Args().Len() == 0 {
		return "", nil
	}
	kind := strings.TrimSpace(cmd.Args().First())
	if err := schemas.ValidateKind(kind); err != nil {
		return "", err
	}
	return kind, nil
}

func parseSchemaGetArguments(cmd *ucli.Command) (string, string, error) {
	if cmd.Args().Len() > 2 {
		return "", "", fmt.Errorf("unexpected args for %s %s", commandSchema, commandSchemaGet)
	}
	if cmd.Args().Len() < 2 {
		return "", "", fmt.Errorf("missing schema declaration type or skel name; expected TYPE SKEL_NAME")
	}
	kind := strings.TrimSpace(cmd.Args().Get(0))
	if err := schemas.ValidateKind(kind); err != nil {
		return "", "", err
	}
	skelName := strings.TrimSpace(cmd.Args().Get(1))
	if skelName == "" {
		return "", "", fmt.Errorf("missing skel name")
	}
	return kind, skelName, nil
}

func filterSchemaEntries(entries []*schemas.Entry, kind string) []*schemas.Entry {
	if kind == "" {
		return entries
	}
	filtered := make([]*schemas.Entry, 0, len(entries))
	for _, entry := range entries {
		if entry.Kind == schemas.DeclarationType(kind) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func loadQuerySchema(cmd *ucli.Command) (*schemas.Document, error) {
	result, err := skelapi.QuerySchema(skelapi.Input{SkelIn: cmd.String(flagSchemaSkelIn), Strict: cmd.Bool(flagStrict)}, skelapi.SchemaQueryOption{})
	if err != nil {
		return nil, generationCommandFailure(err)
	}
	writeWarningLogs(cmd, result.Diagnostics)
	return result.Document, nil
}

func loadSourceSchema(cmd *ucli.Command, flagName, skelIn string) (*schemas.Document, error) {
	if strings.TrimSpace(skelIn) == "" {
		return nil, commandFailure(output.ErrorCodeInvalidArgument, fmt.Errorf("missing flag %s", flagName))
	}
	result, err := skelapi.QuerySchema(skelapi.Input{SkelIn: skelIn, Strict: cmd.Bool(flagStrict)}, skelapi.SchemaQueryOption{})
	if err != nil {
		return nil, generationCommandFailure(err)
	}
	writeWarningLogs(cmd, result.Diagnostics)
	return result.Document, nil
}

func writeSchemaResult(cmd *ucli.Command, value any, context string) error {
	if err := writeJSONResult(cmd, value, context); err != nil {
		return commandFailure(output.ErrorCodeCommandFailed, err)
	}
	return nil
}
