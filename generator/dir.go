package generator

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/efritz/go-genlib/command"
	"github.com/efritz/go-genlib/paths"
	"github.com/efritz/go-genlib/types"
)

func generateDirectory(
	appName string,
	ifaces []*types.Interface,
	opts *command.Options,
	filenameGenerator FilenameGenerator,
	interfaceGenerator InterfaceGenerator,
) error {
	dirname := filepath.Join(opts.OutputDir, opts.OutputFilename)

	if !opts.Force {
		allPaths := []string{}
		for _, iface := range ifaces {
			allPaths = append(allPaths, getFilename(
				dirname,
				iface.Name,
				opts.Prefix,
				filenameGenerator,
			))
		}

		conflict, err := paths.AnyExists(allPaths)
		if err != nil {
			return err
		}

		if conflict != "" {
			return fmt.Errorf("filename %s already exists", conflict)
		}
	}

	for _, iface := range ifaces {
		content, err := generateContent(
			appName,
			[]*types.Interface{iface},
			opts.PkgName,
			opts.Prefix,
			interfaceGenerator,
		)

		if err != nil {
			return err
		}

		filename := getFilename(
			dirname,
			iface.Name,
			opts.Prefix,
			filenameGenerator,
		)

		if err := writeFile(filename, content); err != nil {
			return err
		}
	}

	return nil
}

func getFilename(dirname, interfaceName, prefix string, filenameGenerator FilenameGenerator) string {
	filename := filenameGenerator(interfaceName)
	if prefix != "" {
		filename = fmt.Sprintf("%s_%s", prefix, filename)
	}

	return path.Join(dirname, strings.Replace(strings.ToLower(filename), "-", "_", -1))
}
