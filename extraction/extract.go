package extraction

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	gotypes "go/types"
	"log"
	"os"
	"path"
	"strings"

	"github.com/efritz/go-genlib/paths"
	"github.com/efritz/go-genlib/types"
)

type Extractor struct {
	workingDirectory string
	fset             *token.FileSet
	typeConfig       gotypes.Config
}

func NewExtractor() (*Extractor, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory (%s)", err.Error())
	}

	return &Extractor{
		workingDirectory: workingDirectory,
		fset:             token.NewFileSet(),
		typeConfig:       gotypes.Config{Importer: importer.For("source", nil)},
	}, nil
}

func (e *Extractor) Extract(importPaths []string) (*types.Packages, error) {
	packages := map[string]*types.Package{}
	for _, importPath := range importPaths {
		path, err := findPath(e.workingDirectory, importPath)
		if err != nil {
			return nil, err
		}

		log.Printf(
			"parsing package '%s'\n",
			paths.GetRelativePath(path),
		)

		pkg, pkgType, err := e.importPath(path, importPath)
		if err != nil {
			return nil, err
		}

		visitor := newVisitor(importPath, pkgType)
		for _, file := range pkg.Files {
			ast.Walk(visitor, file)
		}

		packages[importPath] = types.NewPackage(importPath, visitor.types)
	}

	return types.NewPackages(packages), nil
}

func (e *Extractor) importPath(path, importPath string) (*ast.Package, *gotypes.Package, error) {
	pkgs, err := parser.ParseDir(e.fset, path, fileFilter, 0)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"could not import package '%s' (%s)",
			importPath,
			err.Error(),
		)
	}

	files := []*ast.File{}
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			files = append(files, file)
		}
	}

	pkgType, err := e.typeConfig.Check("", e.fset, files, nil)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"could not import package '%s' (%s)",
			importPath,
			err.Error(),
		)
	}

	if pkg := getFirst(pkgs); pkg != nil {
		return pkg, pkgType, nil
	}

	return nil, nil, fmt.Errorf(
		"could not import package '%s' (no files in import path)",
		importPath,
	)
}

func fileFilter(info os.FileInfo) bool {
	var (
		name = info.Name()
		ext  = path.Ext(name)
		base = strings.TrimSuffix(name, ext)
	)

	return !info.IsDir() && ext == ".go" && !strings.HasSuffix(base, "_test")
}

func getFirst(pkgs map[string]*ast.Package) *ast.Package {
	if len(pkgs) == 1 {
		for _, pkg := range pkgs {
			return pkg
		}
	}

	return nil
}
