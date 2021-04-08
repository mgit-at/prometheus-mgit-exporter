package nogofmt

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "gofmt",
	Doc:  "check if files are properly formatted",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		filename := pass.Fset.Position(file.Pos()).Filename
		if ignoreFile(filename) {
			continue
		}

		in, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("ioutil.ReadFile %q: %w", filename, err)
		}

		out, err := format.Source(in)
		if err != nil {
			return nil, fmt.Errorf("format.Source %q: %w", filename, err)
		}

		if bytes.Equal(in, out) {
			continue
		}

		pass.Reportf(file.Pos(), "file is incorrectly formatted; please run `go fmt`")
	}
	return nil, nil
}

func ignoreFile(filename string) bool {
	base := filepath.Base(filename)
	return base == "_cgo_gotypes.go" ||
		base == "_cgo_imports.go" ||
		base == "testmain.go" ||
		strings.Contains(filename, "/external/")
}
