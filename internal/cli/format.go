package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	ucli "github.com/urfave/cli/v3"
	"go.yorun.ai/skelc"
	"go.yorun.ai/skelc/internal/command"
	"go.yorun.ai/skelc/internal/util/fileutil"
)

const (
	commandFormat = "format"

	flagFormatSkelIn = "skel-in"
	flagFormatCheck  = "check"
)

type _FormatOption struct {
	skelIn string
	check  bool
	strict bool
}

type _FormatResult = command.FormatResult

func newFormatCommand() *ucli.Command {
	return &ucli.Command{
		Name:  commandFormat,
		Usage: "format skel definition files",
		Flags: []ucli.Flag{
			&ucli.StringFlag{Name: flagFormatSkelIn, Usage: "skeleton input file or directory"},
			&ucli.BoolFlag{Name: flagFormatCheck, Usage: "check formatting without modifying files"},
		},
		Action: func(_ context.Context, cmd *ucli.Command) error {
			option, err := parseFormatCommand(cmd)
			if err != nil {
				return commandFailure(command.ErrorCodeInvalidArgument, err)
			}
			result, err := formatFiles(option)
			if err != nil {
				code := command.ErrorCodeCommandFailed
				if errors.Is(err, skelc.ErrFormatCompilation) {
					code = command.ErrorCodeCompilationFailed
				}
				return commandFailure(code, err)
			}
			if err := writeFormatResult(cmd, result); err != nil {
				return commandFailure(command.ErrorCodeCommandFailed, err)
			}
			if option.check && result.Changed {
				return commandUnsatisfied()
			}
			return nil
		},
	}
}

func parseFormatCommand(cmd *ucli.Command) (_FormatOption, error) {
	if cmd.Args().Len() != 0 {
		return _FormatOption{}, fmt.Errorf("unexpected args for %s", commandFormat)
	}
	skelIn := cmd.String(flagFormatSkelIn)
	if skelIn == "" {
		return _FormatOption{}, fmt.Errorf("missing flag skel-in")
	}
	path, err := filepath.Abs(skelIn)
	if err != nil {
		return _FormatOption{}, fmt.Errorf("resolve path %s: %w", skelIn, err)
	}
	return _FormatOption{skelIn: path, check: cmd.Bool(flagFormatCheck), strict: cmd.Bool(flagStrict)}, nil
}

func formatFiles(option _FormatOption) (_FormatResult, error) {
	planned, err := skelc.FormatFiles(skelc.FormatOption{SkelIn: option.skelIn, Strict: option.strict})
	if err != nil {
		return _FormatResult{}, err
	}
	formattedFiles := planned.Files
	result := _FormatResult{Changed: len(formattedFiles) > 0, Files: make([]string, 0, len(formattedFiles))}
	for _, file := range formattedFiles {
		result.Files = append(result.Files, file.Path)
	}
	if option.check {
		return result, nil
	}
	replacements := make([]fileutil.Replacement, 0, len(formattedFiles))
	for _, file := range formattedFiles {
		writePath, err := filepath.EvalSymlinks(file.Path)
		if err != nil {
			return _FormatResult{}, fmt.Errorf("resolve format target %s: %w", file.Path, err)
		}
		info, err := os.Stat(writePath)
		if err != nil {
			return _FormatResult{}, fmt.Errorf("inspect format target %s: %w", writePath, err)
		}
		current, err := os.ReadFile(writePath)
		if err != nil {
			return _FormatResult{}, fmt.Errorf("read format target %s: %w", writePath, err)
		}
		if !bytes.Equal(current, file.Original) {
			return _FormatResult{}, fmt.Errorf("format target %s changed while formatting", writePath)
		}
		replacements = append(replacements, fileutil.Replacement{
			Path: writePath, Content: file.Content, Mode: info.Mode(),
		})
	}
	if err := fileutil.ReplaceAll(replacements); err != nil {
		return _FormatResult{}, err
	}
	return result, nil
}

func writeFormatResult(cmd *ucli.Command, result _FormatResult) error {
	return writeJSONResult(cmd, result, "format result")
}
