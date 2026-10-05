package cli

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	ucli "github.com/urfave/cli/v3"
	"go.yorun.ai/skelc/internal/command"
	"go.yorun.ai/skelc/internal/compiler"
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
				return commandFailure(command.ErrorCodeInvalidArgument, fmt.Errorf("unexpected args for scan imports"))
			}
			option := compiler.Option{SkelIn: cmd.String(flagScanSkelIn), Strict: cmd.Bool(flagStrict)}
			if err := normalizeCompilerOption(&option); err != nil {
				return commandFailure(command.ErrorCodeInvalidArgument, err)
			}
			result, err := compiler.CompileImportContext(ctx, option)
			if err != nil {
				return commandFailure(command.ErrorCodeCompilationFailed, err)
			}
			writeWarningLogs(cmd, result.Diagnostics)
			imports := make([]command.ScanImport, 0, len(result.Imports))
			for _, imported := range result.Imports {
				imports = append(imports, command.ScanImport{
					Domain: imported.Name, Alias: imported.Alias,
					File: imported.Pos.File, Line: imported.Pos.Line, Column: imported.Pos.Column,
				})
			}
			slices.SortFunc(imports, func(a, b command.ScanImport) int {
				return cmp.Or(cmp.Compare(a.File, b.File), cmp.Compare(a.Line, b.Line), cmp.Compare(a.Column, b.Column))
			})
			return writeSchemaResult(cmd, imports, "source imports")
		},
	}
}
