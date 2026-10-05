package cli

import (
	"context"
	"fmt"

	ucli "github.com/urfave/cli/v3"
	"go.yorun.ai/skelc"
	"go.yorun.ai/skelc/internal/command"
)

func newSchemaDepCommand() *ucli.Command {
	return &ucli.Command{Name: "dep", Usage: "query complete, public, or API declarations and external dependencies", Flags: []ucli.Flag{
		&ucli.BoolFlag{Name: flagGenPub, Usage: "query backend public contract dependencies"},
		&ucli.BoolFlag{Name: flagGenApi, Usage: "query API dependencies"},
		&ucli.BoolFlag{Name: flagGenPrune, Usage: "retain only selected API roots and their type dependencies; requires --api"},
		&ucli.StringSliceFlag{Name: flagGenActor, Usage: "fully qualified API actor name; requires --api, repeat to select multiple actors"},
		&ucli.StringSliceFlag{Name: flagGenName, Usage: "fully qualified local data or enum root; requires --api --prune, repeat to select multiple types"},
		&ucli.StringFlag{Name: flagGenSkelIn, Usage: "skeleton input file or directory"},
		&ucli.StringSliceFlag{Name: flagGenSkelImport, Usage: "skel dependency mapping in domain=path form; repeat for transitive imports"},
	}, Action: func(_ context.Context, cmd *ucli.Command) error {
		if cmd.Args().Len() != 0 {
			return commandFailure(command.ErrorCodeInvalidArgument, fmt.Errorf("unexpected args for schema dep"))
		}
		imports, err := parseMappingFlags(cmd.StringSlice(flagGenSkelImport), flagGenSkelImport)
		if err != nil {
			return commandFailure(command.ErrorCodeInvalidArgument, err)
		}
		result, err := skelc.QuerySchemaDependencies(skelc.Input{SkelIn: cmd.String(flagGenSkelIn), SkelImports: imports, Strict: cmd.Bool(flagStrict)}, skelc.SchemaDependencyOption{Pub: cmd.Bool(flagGenPub), Api: cmd.Bool(flagGenApi), ApiFilter: skelc.ApiFilter{Actors: cmd.StringSlice(flagGenActor), Prune: cmd.Bool(flagGenPrune), Types: cmd.StringSlice(flagGenName)}})
		if err != nil {
			return generationCommandFailure(err)
		}
		writeWarningLogs(cmd, result.Diagnostics)
		return writeSchemaResult(cmd, result.Report, "schema dependencies")
	}}
}
