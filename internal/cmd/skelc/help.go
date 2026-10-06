package skelc

import (
	"fmt"
	"io"
	"strings"

	ucli "github.com/urfave/cli/v3"
)

const groupCommandHelpTemplate = `NAME:
   {{template "helpNameTemplate" .}}

USAGE:
   {{if .UsageText}}{{wrap .UsageText 3}}{{else}}{{.FullName}}{{if .VisibleCommands}} [command [command options]]{{end}}{{if .ArgsUsage}} {{.ArgsUsage}}{{else}}{{if .Arguments}} [arguments...]{{end}}{{end}}{{end}}{{if .Category}}

CATEGORY:
   {{.Category}}{{end}}{{if .Description}}

DESCRIPTION:
   {{template "descriptionTemplate" .}}{{end}}{{if .VisibleCommands}}

COMMANDS:{{ $cv := offsetCommandsWithArgs .VisibleCommands 5}}{{range .VisibleCommands}}
   {{$s := commandNamesWithArgs .}}{{$s}}{{ $sp := subtract $cv (offset $s 3) }}{{ indent $sp ""}}{{wrap .Usage $cv}}{{end}}{{end}}
{{- range .VisibleCommands}}{{if .VisibleFlags}}

{{.Name}} OPTIONS:{{template "visibleFlagTemplate" .}}{{end}}{{end}}
{{- if .VisibleFlags}}

OPTIONS:{{template "visibleFlagTemplate" .}}{{end}}
`

func init() {
	if helpFlag, ok := ucli.HelpFlag.(*ucli.BoolFlag); ok {
		helpFlag.Hidden = true
	}

	ucli.HelpPrinter = func(writer io.Writer, tpl string, data any) {
		ucli.HelpPrinterCustom(writer, tpl, data, map[string]any{
			"commandNamesWithArgs":   commandNamesWithArgs,
			"offsetCommandsWithArgs": offsetCommandsWithArgs,
		})
	}

	ucli.ShowSubcommandHelp = func(cmd *ucli.Command) error {
		tpl := cmd.CustomHelpTemplate
		if tpl == "" {
			tpl = ucli.SubcommandHelpTemplate
		}
		ucli.HelpPrinter(cmd.Root().Writer, tpl, cmd)
		return nil
	}
}

func commandNamesWithArgs(cmd *ucli.Command) string {
	names := strings.Join(cmd.Names(), ", ")
	if cmd.ArgsUsage == "" {
		return names
	}
	return fmt.Sprintf("%s %s", names, cmd.ArgsUsage)
}

func offsetCommandsWithArgs(cmds []*ucli.Command, fixed int) int {
	maxLen := 0
	for _, cmd := range cmds {
		if length := len(commandNamesWithArgs(cmd)); length > maxLen {
			maxLen = length
		}
	}
	return maxLen + fixed
}
