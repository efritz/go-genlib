package paths

import (
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	srcpath       = filepath.Join(Gopath(), "src")
	modulePattern = regexp.MustCompile(`^module\s+(.+)$`)
)

func InferImportPath(dirname string) (string, bool) {
	if module, wd, ok := Module(dirname); ok {
		return filepath.Join(module, dirname[len(wd):]), true
	}

	if strings.HasPrefix(dirname, srcpath) {
		return dirname[len(srcpath):], true
	}

	return "", false
}

func ResolveImportPath(wd, importPath string) (string, string, bool) {
	// See if we're in a module and generating for our own package
	if module, baseDir, ok := Module(wd); ok && strings.HasPrefix(importPath, module) {
		return importPath, filepath.Join(baseDir, importPath[len(module):]), true
	}

	// See if it's a relative path to working directory
	if dir := filepath.Join(wd, importPath); DirExists(dir) {
		if path, ok := InferImportPath(dir); ok {
			return path, dir, true
		}
	}

	if strings.HasPrefix(wd, srcpath) {
		for wd != srcpath {
			// See if it's vendored on any path up to the GOPATH root
			if dir := filepath.Join(wd, "vendor", importPath); DirExists(dir) {
				return importPath, dir, true
			}

			wd = filepath.Dir(wd)
		}
	}

	// See if it's in the GOPATH
	if dir := filepath.Join(srcpath, importPath); DirExists(dir) {
		return importPath, dir, true
	}

	return "", "", false
}

func Module(dirname string) (string, string, bool) {
	wd := dirname
	for wd != srcpath && wd != "/" {
		if module, ok := Gomod(wd); ok {
			return module, wd, true
		}

		wd = filepath.Dir(wd)
	}

	return "", "", false
}

func Gomod(dirname string) (string, bool) {
	content, err := ioutil.ReadFile(filepath.Join(dirname, "go.mod"))
	if err != nil {
		return "", false
	}

	for _, line := range strings.Split(string(content), "\n") {
		if matches := modulePattern.FindStringSubmatch(line); len(matches) > 0 {
			return matches[1], true
		}
	}

	return "", false
}

func Gopath() string {
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		return gopath
	}

	return build.Default.GOPATH
}
