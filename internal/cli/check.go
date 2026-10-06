package cli

import (
	"context"
	"fmt"

	ucli "github.com/urfave/cli/v3"
	"go.yorun.ai/skelc"
	"go.yorun.ai/skelc/diagnostic"
	"go.yorun.ai/skelc/internal/command"
)

const (
	commandCheck = "check"

	flagCheckSkelIn = "skel-in"
)

func newCheckCommand() *ucli.Command {
	return &ucli.Command{
		Name:  commandCheck,
		Usage: "validate skel definition files",
		Flags: []ucli.Flag{
			&ucli.StringFlag{Name: flagCheckSkelIn, Usage: "skeleton input file or directory"},
		},
		Action: func(ctx context.Context, cmd *ucli.Command) error {
			option, err := parseCheckCommand(cmd)
			if err != nil {
				return commandFailure(command.ErrorCodeInvalidArgument, err)
			}
			result, err := skelc.CheckContext(ctx, option)
			if err != nil {
				return commandFailure(command.ErrorCodeCompilationFailed, err)
			}
			diagnostics := result.Diagnostics
			if diagnostics == nil {
				diagnostics = []diagnostic.Diagnostic{}
			}
			valid := result.Valid
			if err := writeJSONResult(cmd, command.CheckResult{Valid: valid, Diagnostics: diagnostics}, "check result"); err != nil {
				return commandFailure(command.ErrorCodeCommandFailed, err)
			}
			if !valid {
				return commandUnsatisfied()
			}
			return nil
		},
	}
}

func parseCheckCommand(cmd *ucli.Command) (skelc.CheckOption, error) {
	if cmd.Args().Len() != 0 {
		return skelc.CheckOption{}, fmt.Errorf("unexpected args for %s", commandCheck)
	}
	option := skelc.CheckOption{
		SkelIn: cmd.String(flagCheckSkelIn),
		Strict: cmd.Bool(flagStrict),
	}
	if option.SkelIn == "" {
		return skelc.CheckOption{}, fmt.Errorf("missing flag skel-in")
	}
	return option, nil
}
