package cli

import (
	"context"
	"fmt"

	ucli "github.com/urfave/cli/v3"
	"go.yorun.ai/skelc"
	"go.yorun.ai/skelc/internal/command"
)

func newSchemaDepCommand() *ucli.Command {
	return &ucli.Command{Name: "dep", Usage: "query selected API declarations and external type dependencies", Flags: []ucli.Flag{
		&ucli.BoolFlag{Name: flagGenApi, Usage: "query API dependencies (required)"},
		&ucli.BoolFlag{Name: flagGenPrune, Usage: "retain only selected API roots and their type dependencies"},
		&ucli.StringSliceFlag{Name: flagGenActor, Usage: "fully qualified API actor name; repeat to select multiple actors"},
		&ucli.StringSliceFlag{Name: flagGenType, Usage: "fully qualified local data or enum root; requires --prune, repeat to select multiple types"},
		&ucli.StringFlag{Name: flagGenSkelIn, Usage: "skeleton input file or directory"},
		&ucli.StringSliceFlag{Name: flagGenSkelImport, Usage: "skel dependency mapping in domain=path form; repeat for transitive imports"},
	}, Action: func(_ context.Context, cmd *ucli.Command) error {
		if cmd.Args().Len() != 0 {
			return commandFailure(command.ErrorCodeInvalidArgument, fmt.Errorf("unexpected args for schema dep"))
		}
		if !cmd.Bool(flagGenApi) {
			return commandFailure(command.ErrorCodeInvalidArgument, fmt.Errorf("flag api is required for schema dep"))
		}
		imports, err := parseMappingFlags(cmd.StringSlice(flagGenSkelImport), flagGenSkelImport)
		if err != nil {
			return commandFailure(command.ErrorCodeInvalidArgument, err)
		}
		result, err := skelc.QueryApiDependencies(skelc.Input{SkelIn: cmd.String(flagGenSkelIn), SkelImports: imports, Strict: cmd.Bool(flagStrict)}, skelc.ApiFilter{Actors: cmd.StringSlice(flagGenActor), Prune: cmd.Bool(flagGenPrune), Types: cmd.StringSlice(flagGenType)})
		if err != nil {
			return generationCommandFailure(err)
		}
		writeWarningLogs(cmd, result.Diagnostics)
		return writeSchemaResult(cmd, result.Report, "API dependencies")
	}}
}
