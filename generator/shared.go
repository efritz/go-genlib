package generator

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/dave/jennifer/jen"
	"github.com/efritz/go-genlib/paths"
	"github.com/efritz/go-genlib/types"
)

func inferImportPath(path string) (string, error) {
	gopath := paths.Gopath()
	if strings.HasPrefix(path, gopath) {
		// gopath + /src/
		return path[len(gopath)+5:], nil
	}

	return "", fmt.Errorf("destination is outside $GOPATH")
}

func writeFile(filename, content string) error {
	return ioutil.WriteFile(filename, []byte(content), 0644)
}

func generateContent(
	appName string,
	ifaces []*types.Interface,
	pkgName string,
	prefix string,
	interfaceGenerator InterfaceGenerator,
) (string, error) {
	file := newFile(appName, pkgName)

	for _, iface := range ifaces {
		interfaceGenerator(file, iface, prefix)
	}

	buffer := &bytes.Buffer{}
	if err := file.Render(buffer); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func newFile(appName, pkgName string) *jen.File {
	file := jen.NewFile(pkgName)
	file.HeaderComment(fmt.Sprintf("Code generated by %s; DO NOT EDIT.", appName))
	file.HeaderComment("This file was generated by robots at")
	file.HeaderComment(time.Now().Format(time.RFC3339))
	file.HeaderComment("using the command")
	file.HeaderComment(fmt.Sprintf("$ %s", strings.Join(os.Args, " ")))

	return file
}