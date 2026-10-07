package skelc

import (
	"context"
	"fmt"
	"strings"

	ucli "github.com/urfave/cli/v3"
	skelapi "go.yorun.ai/skel/api"
	"go.yorun.ai/skel/internal/cmd/skelc/output"
)

const (
	commandGen         = "gen"
	commandGenGo       = "go"
	commandGenGoModule = "go-module"
	commandGenSkel     = "skel"
	commandGenTS       = "ts"

	flagGenPub            = "pub"
	flagGenSkelIn         = "skel-in"
	flagGenSkelOut        = "skel-out"
	flagGenGoOut          = "go-out"
	flagGenGoPubOut       = "go-pub-out"
	flagGenTSOut          = "ts-out"
	flagGenApi            = "api"
	flagGenActor          = "actor"
	flagGenPrune          = "prune"
	flagGenName           = "name"
	flagGenGoVrpcVersion  = "go-vrpc-version"
	flagGenGoModulePrefix = "go-module-prefix"
	flagGenGoModule       = "go-module"
	flagGenGoPubModule    = "go-pub-module"
	flagGenGoImport       = "go-import"
	flagGenTSAsModule     = "ts-as-module"
	flagGenTSModuleScope  = "ts-module-scope"
	flagGenTSModule       = "ts-module"
	flagGenTSImport       = "ts-import"
	flagGenSkelImport     = "skel-import"
	flagGenGoVineVersion  = "go-vine-version"
)

func newGenCommand() *ucli.Command {
	return &ucli.Command{
		Name:               commandGen,
		Usage:              "generate code from skel definitions",
		HideHelpCommand:    true,
		CustomHelpTemplate: groupCommandHelpTemplate,
		Commands: []*ucli.Command{
			newGenSkelCommand(),
			newGenGoCommand(),
			newGenGoModuleCommand(),
			newGenTSCommand(),
		},
	}
}

func newGenGoCommand() *ucli.Command {
	return &ucli.Command{
		Name:  commandGenGo,
		Usage: "generate non-module Go code from skel definitions",
		Flags: newGenGoFlags(),
		Action: func(_ context.Context, cmd *ucli.Command) error {
			input, option, err := parseGenGoCommand(cmd)
			if err != nil {
				return commandFailure(output.ErrorCodeInvalidArgument, err)
			}
			option.CompilerVersion, err = compilerVersion()
			if err != nil {
				return commandFailure(output.ErrorCodeCommandFailed, err)
			}
			result, err := skelapi.CompileGolang(input, option)
			if err != nil {
				return generationCommandFailure(err)
			}
			return writeGenerationResult(cmd, result)
		},
	}
}

func newGenGoModuleCommand() *ucli.Command {
	return &ucli.Command{
		Name:  commandGenGoModule,
		Usage: "generate Go module code from skel definitions",
		Flags: newGenGoModuleFlags(),
		Action: func(_ context.Context, cmd *ucli.Command) error {
			input, option, err := parseGenGoModuleCommand(cmd)
			if err != nil {
				return commandFailure(output.ErrorCodeInvalidArgument, err)
			}
			option.CompilerVersion, err = compilerVersion()
			if err != nil {
				return commandFailure(output.ErrorCodeCommandFailed, err)
			}
			result, err := skelapi.CompileGolang(input, option)
			if err != nil {
				return generationCommandFailure(err)
			}
			return writeGenerationResult(cmd, result)
		},
	}
}

func newGenTSCommand() *ucli.Command {
	return &ucli.Command{
		Name:  commandGenTS,
		Usage: "generate TypeScript code from skel definitions",
		Flags: newGenTSFlags(),
		Action: func(_ context.Context, cmd *ucli.Command) error {
			input, option, err := parseGenTSCommand(cmd)
			if err != nil {
				return commandFailure(output.ErrorCodeInvalidArgument, err)
			}
			result, err := skelapi.CompileTypeScript(input, option)
			if err != nil {
				return generationCommandFailure(err)
			}
			return writeGenerationResult(cmd, result)
		},
	}
}

func newGenSkelCommand() *ucli.Command {
	return &ucli.Command{
		Name:  commandGenSkel,
		Usage: "generate pub skel definitions from skel definitions",
		Flags: newGenSkelFlags(),
		Action: func(_ context.Context, cmd *ucli.Command) error {
			input, option, err := parseGenSkelCommand(cmd)
			if err != nil {
				return commandFailure(output.ErrorCodeInvalidArgument, err)
			}
			result, err := skelapi.CompileSkeleton(input, option)
			if err != nil {
				return generationCommandFailure(err)
			}
			return writeGenerationResult(cmd, result)
		},
	}
}

func writeGenerationResult(cmd *ucli.Command, result skelapi.CompileResult) error {
	writeWarningLogs(cmd, result.Diagnostics)
	if err := writeJSONResult(cmd, output.GenerationResult{Generated: true}, "generation result"); err != nil {
		return commandFailure(output.ErrorCodeCommandFailed, err)
	}
	return nil
}

func compilerVersion() (string, error) {
	info, err := debugBuildInfo()
	return info.Version, err
}

func newGenGoFlags() []ucli.Flag {
	return []ucli.Flag{
		&ucli.BoolFlag{Name: flagGenApi, Usage: "generate portal API clients"},
		&ucli.StringSliceFlag{Name: flagGenActor, Usage: "fully qualified API actor name; requires --api, repeat to select multiple actors"},
		&ucli.BoolFlag{Name: flagGenPrune, Usage: "retain only selected API roots and their type dependencies; requires --api"},
		&ucli.StringSliceFlag{Name: flagGenName, Usage: "fully qualified local data or enum root; requires --api --prune, repeat to select multiple types"},
		&ucli.BoolFlag{Name: flagGenPub, Usage: "generate backend public contracts"},
		&ucli.StringFlag{Name: flagGenGoVrpcVersion, Usage: "vRPC module version for API clients"},
		&ucli.StringFlag{Name: flagGenSkelIn, Usage: "skeleton input file or directory"},
		&ucli.StringFlag{Name: flagGenGoOut, Usage: "Go output directory"},
		&ucli.StringSliceFlag{Name: flagGenSkelImport, Usage: "Skel dependency in domain=path form"},
		&ucli.StringSliceFlag{Name: flagGenGoImport, Usage: "Go dependency in domain=package form"},
		&ucli.StringFlag{Name: flagGenGoVineVersion, Usage: "Vine module version for generated Go code"},
	}
}

func parseGenGoCommand(cmd *ucli.Command) (skelapi.Input, skelapi.GolangOption, error) {
	if cmd.Args().Len() != 0 {
		return skelapi.Input{}, skelapi.GolangOption{}, fmt.Errorf("unexpected args for %s %s", commandGen, commandGenGo)
	}

	skelImports, err := parseMappingFlags(cmd.StringSlice(flagGenSkelImport), flagGenSkelImport)
	if err != nil {
		return skelapi.Input{}, skelapi.GolangOption{}, err
	}
	goImports, err := parseMappingFlags(cmd.StringSlice(flagGenGoImport), flagGenGoImport)
	if err != nil {
		return skelapi.Input{}, skelapi.GolangOption{}, err
	}
	input := skelapi.Input{
		SkelIn:      cmd.String(flagGenSkelIn),
		SkelImports: skelImports,
		Strict:      cmd.Bool(flagStrict),
	}

	option := skelapi.GolangOption{
		ApiFilter:   skelapi.ApiFilter{Actors: cmd.StringSlice(flagGenActor), Prune: cmd.Bool(flagGenPrune), Types: cmd.StringSlice(flagGenName)},
		Imports:     goImports,
		PubOnly:     cmd.Bool(flagGenPub),
		ApiOnly:     cmd.Bool(flagGenApi),
		VrpcVersion: cmd.String(flagGenGoVrpcVersion),
		Out:         cmd.String(flagGenGoOut),
		VineVersion: strings.TrimSpace(cmd.String(flagGenGoVineVersion)),
	}
	return input, option, nil
}

func newGenGoModuleFlags() []ucli.Flag {
	return []ucli.Flag{
		&ucli.BoolFlag{Name: flagGenApi, Usage: "generate portal API clients"},
		&ucli.StringSliceFlag{Name: flagGenActor, Usage: "fully qualified API actor name; requires --api, repeat to select multiple actors"},
		&ucli.BoolFlag{Name: flagGenPrune, Usage: "retain only selected API roots and their type dependencies; requires --api"},
		&ucli.StringSliceFlag{Name: flagGenName, Usage: "fully qualified local data or enum root; requires --api --prune, repeat to select multiple types"},
		&ucli.BoolFlag{Name: flagGenPub, Usage: "generate backend public contracts"},
		&ucli.StringFlag{Name: flagGenGoVrpcVersion, Usage: "vRPC module version for API clients"},
		&ucli.StringFlag{Name: flagGenSkelIn, Usage: "skeleton input file or directory"},
		&ucli.StringSliceFlag{Name: flagGenSkelImport, Usage: "skel dependency mapping in domain=path form; repeat for transitive imports"},
		&ucli.StringFlag{Name: flagGenGoOut, Usage: "Go output directory"},
		&ucli.StringFlag{Name: flagGenGoModule, Usage: "Go module path"},
		&ucli.StringFlag{Name: flagGenGoPubOut, Usage: "Go pub output directory"},
		&ucli.StringFlag{Name: flagGenGoPubModule, Usage: "Go pub module path"},
		&ucli.StringSliceFlag{Name: flagGenGoImport, Usage: "Go import mapping in domain=package form"},
		&ucli.StringFlag{Name: flagGenGoModulePrefix, Usage: "Go module path prefix"},
		&ucli.StringFlag{Name: flagGenGoVineVersion, Usage: "Vine module version for generated Go modules"},
	}
}

func parseGenGoModuleCommand(cmd *ucli.Command) (skelapi.Input, skelapi.GolangOption, error) {
	if cmd.Args().Len() != 0 {
		return skelapi.Input{}, skelapi.GolangOption{}, fmt.Errorf("unexpected args for %s %s", commandGen, commandGenGoModule)
	}

	skelImports, err := parseMappingFlags(cmd.StringSlice(flagGenSkelImport), flagGenSkelImport)
	if err != nil {
		return skelapi.Input{}, skelapi.GolangOption{}, err
	}
	goImports, err := parseMappingFlags(cmd.StringSlice(flagGenGoImport), flagGenGoImport)
	if err != nil {
		return skelapi.Input{}, skelapi.GolangOption{}, err
	}

	input := skelapi.Input{
		SkelIn:      cmd.String(flagGenSkelIn),
		SkelImports: skelImports,
		Strict:      cmd.Bool(flagStrict),
	}

	option := skelapi.GolangOption{
		ApiFilter:    skelapi.ApiFilter{Actors: cmd.StringSlice(flagGenActor), Prune: cmd.Bool(flagGenPrune), Types: cmd.StringSlice(flagGenName)},
		PubOnly:      cmd.Bool(flagGenPub),
		ApiOnly:      cmd.Bool(flagGenApi),
		VrpcVersion:  cmd.String(flagGenGoVrpcVersion),
		Out:          cmd.String(flagGenGoOut),
		AsModule:     true,
		PubOut:       cmd.String(flagGenGoPubOut),
		ModulePrefix: cmd.String(flagGenGoModulePrefix),
		Module:       cmd.String(flagGenGoModule),
		PubModule:    cmd.String(flagGenGoPubModule),
		Imports:      goImports,
		VineVersion:  strings.TrimSpace(cmd.String(flagGenGoVineVersion)),
	}
	return input, option, nil
}

func newGenSkelFlags() []ucli.Flag {
	return []ucli.Flag{
		&ucli.BoolFlag{Name: flagGenPub, Usage: "required, generate only pub skel definitions"},
		&ucli.StringFlag{Name: flagGenSkelIn, Usage: "skeleton input file or directory"},
		&ucli.StringFlag{Name: flagGenSkelOut, Usage: "skel output directory"},
		&ucli.StringSliceFlag{Name: flagGenSkelImport, Usage: "skel dependency mapping in domain=path form; repeat for transitive imports"},
	}
}

func parseGenSkelCommand(cmd *ucli.Command) (skelapi.Input, skelapi.SkeletonOption, error) {
	if cmd.Args().Len() != 0 {
		return skelapi.Input{}, skelapi.SkeletonOption{}, fmt.Errorf("unexpected args for %s %s", commandGen, commandGenSkel)
	}

	skelImports, err := parseMappingFlags(cmd.StringSlice(flagGenSkelImport), flagGenSkelImport)
	if err != nil {
		return skelapi.Input{}, skelapi.SkeletonOption{}, err
	}
	input := skelapi.Input{
		SkelIn:      cmd.String(flagGenSkelIn),
		SkelImports: skelImports,
		Strict:      cmd.Bool(flagStrict),
	}
	option := skelapi.SkeletonOption{
		PubOnly: cmd.Bool(flagGenPub),
		Out:     cmd.String(flagGenSkelOut),
	}
	return input, option, nil
}

func newGenTSFlags() []ucli.Flag {
	return []ucli.Flag{
		&ucli.BoolFlag{Name: flagGenApi, Usage: "generate portal API clients"},
		&ucli.StringSliceFlag{Name: flagGenActor, Usage: "fully qualified API actor name; requires --api, repeat to select multiple actors"},
		&ucli.BoolFlag{Name: flagGenPrune, Usage: "retain only selected API roots and their type dependencies; requires --api"},
		&ucli.StringSliceFlag{Name: flagGenName, Usage: "fully qualified local data or enum root; requires --api --prune, repeat to select multiple types"},
		&ucli.StringFlag{Name: flagGenSkelIn, Usage: "skeleton input file or directory"},
		&ucli.StringFlag{Name: flagGenTSOut, Usage: "TypeScript output directory"},
		&ucli.StringSliceFlag{Name: flagGenSkelImport, Usage: "skel dependency mapping in domain=path form; repeat for transitive imports"},
		&ucli.BoolFlag{Name: flagGenTSAsModule, Usage: "generate TypeScript module package metadata"},
		&ucli.StringFlag{Name: flagGenTSModuleScope, Usage: "TypeScript module package scope"},
		&ucli.StringFlag{Name: flagGenTSModule, Usage: "TypeScript module package name"},
		&ucli.StringSliceFlag{Name: flagGenTSImport, Usage: "TypeScript import mapping in domain=package form"},
	}
}

func parseGenTSCommand(cmd *ucli.Command) (skelapi.Input, skelapi.TypeScriptOption, error) {
	if cmd.Args().Len() != 0 {
		return skelapi.Input{}, skelapi.TypeScriptOption{}, fmt.Errorf("unexpected args for %s %s", commandGen, commandGenTS)
	}

	skelImports, err := parseMappingFlags(cmd.StringSlice(flagGenSkelImport), flagGenSkelImport)
	if err != nil {
		return skelapi.Input{}, skelapi.TypeScriptOption{}, err
	}
	tsImports, err := parseMappingFlags(cmd.StringSlice(flagGenTSImport), flagGenTSImport)
	if err != nil {
		return skelapi.Input{}, skelapi.TypeScriptOption{}, err
	}
	input := skelapi.Input{
		SkelIn:      cmd.String(flagGenSkelIn),
		SkelImports: skelImports,
		Strict:      cmd.Bool(flagStrict),
	}
	option := skelapi.TypeScriptOption{
		ApiFilter:   skelapi.ApiFilter{Actors: cmd.StringSlice(flagGenActor), Prune: cmd.Bool(flagGenPrune), Types: cmd.StringSlice(flagGenName)},
		ApiOnly:     cmd.Bool(flagGenApi),
		Out:         cmd.String(flagGenTSOut),
		AsModule:    cmd.Bool(flagGenTSAsModule),
		ModuleScope: cmd.String(flagGenTSModuleScope),
		Module:      cmd.String(flagGenTSModule),
		Imports:     tsImports,
	}
	return input, option, nil
}
