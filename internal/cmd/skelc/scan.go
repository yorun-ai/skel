package skelc

import (
	"context"
	"fmt"

	ucli "github.com/urfave/cli/v3"
	skelapi "go.yorun.ai/skel/api"
	"go.yorun.ai/skel/internal/cmd/skelc/output"
)

const (
	commandScan        = "scan"
	commandScanImports = "imports"
	flagScanSkelIn     = "skel-in"
)

func newScanCommand() *ucli.Command {
	return &ucli.Command{
		Name:               commandScan,
		Usage:              "scan Skel source without loading dependencies",
		HideHelpCommand:    true,
		CustomHelpTemplate: groupCommandHelpTemplate,
		Commands:           []*ucli.Command{newScanImportsCommand()},
	}
}

func newScanImportsCommand() *ucli.Command {
	return &ucli.Command{
		Name:  commandScanImports,
		Usage: "list direct domain import declarations",
		Flags: []ucli.Flag{&ucli.StringFlag{Name: flagScanSkelIn, Usage: "skeleton input file or directory"}},
		Action: func(ctx context.Context, cmd *ucli.Command) error {
			if cmd.Args().Len() != 0 {
				return commandFailure(output.ErrorCodeInvalidArgument, fmt.Errorf("unexpected args for scan imports"))
			}
			result, err := skelapi.ScanImportsContext(ctx, skelapi.ScanOption{SkelIn: cmd.String(flagScanSkelIn), Strict: cmd.Bool(flagStrict)})
			if err != nil {
				return generationCommandFailure(err)
			}
			writeWarningLogs(cmd, result.Diagnostics)
			imports := make([]output.ScanImport, 0, len(result.Imports))
			for _, imported := range result.Imports {
				imports = append(imports, output.ScanImport(imported))
			}
			return writeSchemaResult(cmd, imports, "source imports")
		},
	}
}
