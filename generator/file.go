package generator

import (
	"fmt"

	"github.com/efritz/go-genlib/command"
	"github.com/efritz/go-genlib/paths"
	"github.com/efritz/go-genlib/types"
)

func generateFile(
	appName string,
	ifaces []*types.Interface,
	opts *command.Options,
	interfaceGenerator InterfaceGenerator,
) error {
	content, err := generateContent(
		appName,
		ifaces,
		opts.PkgName,
		opts.Prefix,
		interfaceGenerator,
	)

	if err != nil {
		return err
	}

	if opts.OutputFilename != "" {
		exists, err := paths.Exists(opts.OutputFilename)
		if err != nil {
			return err
		}

		if exists && !opts.Force {
			return fmt.Errorf("filename %s already exists", opts.OutputFilename)
		}

		return writeFile(opts.OutputFilename, content)
	}

	fmt.Printf("%s\n", content)
	return nil
}
