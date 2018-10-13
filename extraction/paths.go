package extraction

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/efritz/go-genlib/paths"
)

func findPath(wd, importPath string) (string, error) {
	for _, path := range getPossiblePaths(wd, importPath) {
		if paths.DirExists(path) {
			return path, nil
		}
	}

	return "", fmt.Errorf("could not locate package %s", importPath)
}

func getPossiblePaths(wd, importPath string) []string {
	var (
		root       = filepath.Join(paths.Gopath(), "src")
		globalPath = filepath.Join(root, importPath)
	)

	if !strings.HasPrefix(wd, root) {
		return []string{globalPath}
	}

	paths := []string{}
	for wd != root {
		paths = append(paths, filepath.Join(wd, "vendor", importPath))
		wd = filepath.Dir(wd)
	}

	return append(paths, globalPath)
}
