package cli

import (
	"context"
	"io"

	ucli "github.com/urfave/cli/v3"
	"go.yorun.ai/skel/internal/lsp"
)

var serveLSP = lsp.Serve

const commandLSP = "lsp"

func newLSPCommand() *ucli.Command {
	return &ucli.Command{
		Name:  commandLSP,
		Usage: "run the Skel language server over standard input and output",
		Action: func(ctx context.Context, cmd *ucli.Command) error {
			return serveLSP(ctx, commandReader(cmd), cmd.Writer, cmd.Bool(flagStrict))
		},
	}
}

func commandReader(cmd *ucli.Command) io.Reader {
	if cmd.Reader != nil {
		return cmd.Reader
	}
	return cmd.Root().Reader
}
