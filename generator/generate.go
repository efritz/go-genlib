package generator

import (
	"github.com/dave/jennifer/jen"
	"github.com/efritz/go-genlib/command"
	"github.com/efritz/go-genlib/types"
)

type (
	FilenameGenerator  func(ifaceName string) string
	InterfaceGenerator func(file *jen.File, iface *types.Interface, prefix string)
)

func Generate(
	appName string,
	ifaces []*types.Interface,
	opts *command.Options,
	filenameGenerator FilenameGenerator,
	interfaceGenerator InterfaceGenerator,
) error {
	importPath, err := inferImportPath(opts.OutputDir)
	if err != nil {
		return err
	}

	for _, iface := range ifaces {
		if iface.ImportPath == importPath {
			iface.ImportPath = ""
		}
	}

	if opts.OutputFilename == "" && opts.OutputDir != "" {
		return generateDirectory(
			appName,
			ifaces,
			opts,
			filenameGenerator,
			interfaceGenerator,
		)
	}

	return generateFile(appName, ifaces, opts, interfaceGenerator)
}
