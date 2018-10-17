package command

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/alecthomas/kingpin"
	"github.com/efritz/go-genlib/paths"
)

type Options struct {
	ImportPaths    []string
	PkgName        string
	Interfaces     []string
	OutputFilename string
	OutputDir      string
	Prefix         string
	Force          bool
}

var GoIdentifierPattern = regexp.MustCompile("^[A-Za-z]([A-Za-z0-9_]*[A-Za-z])?$")

func parseArgs(
	name string,
	description string,
	version string,
	argHook ArgHookFunc,
	argValidator ArgValidatorFunc,
) (*Options, error) {
	app := kingpin.New(name, description).Version(version)

	opts := &Options{
		ImportPaths: []string{},
		Interfaces:  []string{},
	}

	app.Arg("path", "The import paths used to search for eligible interfaces").Required().StringsVar(&opts.ImportPaths)
	app.Flag("dirname", "The target output directory. Each mock will be written to a unique file.").Short('d').StringVar(&opts.OutputDir)
	app.Flag("filename", "The target output file. All mocks are written to this file.").Short('o').StringVar(&opts.OutputFilename)
	app.Flag("force", "Do not abort if a write to disk would overwrite an existing file.").Short('f').BoolVar(&opts.Force)
	app.Flag("interfaces", "A whitelist of interfaces to generate given the import paths.").Short('i').StringsVar(&opts.Interfaces)
	app.Flag("package", "The name of the generated package. Is the name of target directory if dirname or filename is supplied by default.").Short('p').StringVar(&opts.PkgName)
	app.Flag("prefix", "A prefix used in the name of each mock struct. Should be TitleCase by convention.").StringVar(&opts.Prefix)
	argHook(app)

	if _, err := app.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	validators := []ArgValidatorFunc{
		validateOutputPaths,
		validateOptions,
		argValidator,
	}

	for _, f := range validators {
		if fatal, err := f(opts); err != nil {
			if !fatal {
				kingpin.Fatalf("%s, try --help", err.Error())
			}

			return nil, err
		}
	}

	return opts, nil
}

func validateOptions(opts *Options) (bool, error) {
	if opts.PkgName == "" {
		if opts.OutputDir == "" {
			return false, fmt.Errorf("could not infer package")
		}

		opts.PkgName = path.Base(opts.OutputDir)
	}

	if !GoIdentifierPattern.Match([]byte(opts.PkgName)) {
		return false, fmt.Errorf("illegal package name supplied")
	}

	if opts.Prefix != "" && !GoIdentifierPattern.Match([]byte(opts.Prefix)) {
		kingpin.Fatalf("illegal prefix supplied, try --help")
	}

	return false, nil
}

func validateOutputPaths(opts *Options) (bool, error) {
	wd, err := os.Getwd()
	if err != nil {
		return true, fmt.Errorf("failed to get current directory")
	}

	if opts.OutputFilename == "" && opts.OutputDir == "" {
		opts.OutputDir = wd
	}

	if opts.OutputFilename != "" && opts.OutputDir != "" {
		return false, fmt.Errorf("dirname and filename are mutually exclusive")
	}

	if opts.OutputFilename != "" {
		filename, err := filepath.Abs(opts.OutputFilename)
		if err != nil {
			return true, err
		}

		opts.OutputDir = path.Dir(filename)
		opts.OutputFilename = path.Base(filename)
	}

	dirname, err := filepath.EvalSymlinks(opts.OutputDir)
	if err != nil {
		return true, err
	}

	opts.OutputDir = dirname

	if err := paths.EnsureDirExists(dirname); err != nil {
		return true, fmt.Errorf(
			"failed to make output directory %s: %s",
			dirname,
			err.Error(),
		)
	}

	return false, nil
}
