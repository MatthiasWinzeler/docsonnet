package docsonnet

import (
	"encoding/json"
	"fmt"
	"log"

	_ "embed"

	"github.com/google/go-jsonnet"
)

//go:embed load.libsonnet
var loadLibsonnet string

//go:embed main.libsonnet
var mainDocUtilLibsonnet string

type Opts struct {
	JPath []string
}

// Load extracts and transforms the docsonnet data in `filename`, returning the
// top level docsonnet package.
func Load(filename string, opts Opts) (*Package, error) {
	data, err := Extract(filename, opts)
	if err != nil {
		return nil, err
	}

	return Transform([]byte(data))
}

// Extract parses the Jsonnet file at `filename`, extracting all docsonnet related
// information, exactly as they appear in Jsonnet. Keep in mind this
// representation is usually not suitable for any use, use `Transform` to
// convert it to the familiar docsonnet data model.
func Extract(filename string, opts Opts) ([]byte, error) {
	// setup Jsonnet vm
	vm := jsonnet.MakeVM()
	importer, err := newImporter(opts.JPath)
	if err != nil {
		return nil, err
	}
	vm.Importer(importer)

	// invoke load.libsonnet
	vm.ExtCode("main", fmt.Sprintf(`(import "%s")`, filename))

	data, err := vm.EvaluateSnippet("load.libsonnet", loadLibsonnet)
	if err != nil {
		return nil, err
	}

	return []byte(data), nil
}

// Transform converts the raw result of `Extract` to the actual docsonnet object
// model `*docsonnet.Package`.
func Transform(data []byte) (*Package, error) {
	var d ds
	if err := json.Unmarshal([]byte(data), &d); err != nil {
		log.Fatalln(err)
	}

	p := fastLoad(d)
	return &p, nil
}

// importer wraps jsonnet.FileImporter, to statically provide load.libsonnet,
// bundled with the binary
type importer struct {
	fi   jsonnet.FileImporter
	util jsonnet.Contents
}

func newImporter(paths []string) (*importer, error) {
	return &importer{
		fi:   jsonnet.FileImporter{JPaths: paths},
		util: jsonnet.MakeContents(mainDocUtilLibsonnet),
	}, nil
}

var docUtilPaths = []string{
	"doc-util/main.libsonnet",
	"github.com/jsonnet-libs/docsonnet/doc-util/main.libsonnet",
}

func (i *importer) Import(importedFrom, importedPath string) (contents jsonnet.Contents, foundAt string, err error) {
	for _, p := range docUtilPaths {
		if importedPath == p {
			return i.util, "<internal>", nil
		}
	}

	return i.fi.Import(importedFrom, importedPath)
}
