package generation

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/efritz/go-genlib/command"
	"github.com/efritz/go-genlib/paths"
	"github.com/efritz/go-genlib/types"
)

type (
	FilenameGenerator  func(name string) string
	InterfaceGenerator func(file *jen.File, iface *types.Interface, prefix string)
)

func Generate(
	appName string,
	appVersion string,
	ifaces []*types.Interface,
	opts *command.Options,
	filenameGenerator FilenameGenerator,
	interfaceGenerator InterfaceGenerator,
) error {
	if opts.OutputFilename == "" && opts.OutputDir != "" {
		return generateDirectory(
			appName,
			appVersion,
			ifaces,
			opts,
			filenameGenerator,
			interfaceGenerator,
		)
	}

	return generateFile(appName, appVersion, ifaces, opts, interfaceGenerator)
}

func generateFile(
	appName string,
	appVersion string,
	ifaces []*types.Interface,
	opts *command.Options,
	interfaceGenerator InterfaceGenerator,
) error {
	content, err := generateContent(
		appName,
		appVersion,
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
			return fmt.Errorf(
				"filename %s already exists, overwrite with --force",
				paths.GetRelativePath(opts.OutputFilename),
			)
		}

		return writeFile(opts.OutputFilename, content)
	}

	fmt.Printf("%s\n", content)
	return nil
}

func generateDirectory(
	appName string,
	appVersion string,
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
			return fmt.Errorf(
				"filename %s already exists, overwrite with --force",
				paths.GetRelativePath(conflict),
			)
		}
	}

	for _, iface := range ifaces {
		content, err := generateContent(
			appName,
			appVersion,
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

func generateContent(
	appName string,
	appVersion string,
	ifaces []*types.Interface,
	pkgName string,
	prefix string,
	interfaceGenerator InterfaceGenerator,
) (string, error) {
	file := newFile(appName, appVersion, pkgName)

	for _, iface := range ifaces {
		log.Printf(
			"generating code for interface '%s'\n",
			iface.Name,
		)

		interfaceGenerator(file, iface, prefix)
	}

	buffer := &bytes.Buffer{}
	if err := file.Render(buffer); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func newFile(appName, appVersion, pkgName string) *jen.File {
	file := jen.NewFile(pkgName)
	file.HeaderComment(fmt.Sprintf("Code generated by %s %s; DO NOT EDIT.", appName, appVersion))
	return file
}

func writeFile(filename, content string) error {
	log.Printf(
		"writing to '%s'\n",
		paths.GetRelativePath(filename),
	)

	return ioutil.WriteFile(filename, []byte(content), 0644)
}
