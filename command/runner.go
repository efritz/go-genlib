package command

import (
	"fmt"
	"strings"

	"github.com/alecthomas/kingpin"

	"github.com/efritz/go-genlib/types"
)

type (
	commandConfig struct {
		argHook      ArgHookFunc
		argValidator ArgValidatorFunc
	}

	TypeGetter       func(pkgs *types.Packages, name string) (*types.Interface, error)
	Generator        func(ifaces []*types.Interface, opts *Options) error
	ArgHookFunc      func(app *kingpin.Application)
	ArgValidatorFunc func(opts *Options) (error, bool)
)

func Run(
	name string,
	description string,
	version string,
	typeGetter TypeGetter,
	generator Generator,
	configs ...ConfigFunc,
) error {
	config := &commandConfig{
		argHook:      func(_ *kingpin.Application) {},
		argValidator: func(_ *Options) (error, bool) { return nil, false },
	}

	for _, f := range configs {
		f(config)
	}

	opts, err := parseArgs(
		name,
		description,
		version,
		config.argHook,
		config.argValidator,
	)

	ifaces, err := Extract(
		typeGetter,
		opts.ImportPaths,
		opts.Interfaces,
	)

	if err != nil {
		return err
	}

	if opts.ListOnly {
		for _, iface := range ifaces {
			fmt.Printf("%s\n", iface.Name)
		}

		return nil
	}

	nameMap := map[string]struct{}{}
	for _, t := range ifaces {
		nameMap[strings.ToLower(t.Name)] = struct{}{}
	}

	for _, name := range opts.Interfaces {
		if _, ok := nameMap[strings.ToLower(name)]; !ok {
			return fmt.Errorf("type '%s' not found in supplied import paths", name)
		}
	}

	return generator(ifaces, opts)
}
