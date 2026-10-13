package generator

import (
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/protoc-contrib/protoc-gen-go-aip/internal/generator/output"
	"github.com/protoc-contrib/protoc-gen-go-aip/internal/generator/query"
	"github.com/protoc-contrib/protoc-gen-go-aip/internal/generator/resource"
)

// Generate runs all passes against the plugin request, each writing into the
// one *_aip.pb.go companion of a proto file. The resource pass walks every
// file (so cross-file refs resolve); the query pass adds filter helpers for
// List requests with a filter field. A file neither pass writes to gets no
// companion.
func Generate(plugin *protogen.Plugin, opts *Options) error {
	if opts == nil {
		opts = &Options{}
	}
	files := output.New(plugin)
	if err := resource.Generate(plugin, files, &opts.Resource); err != nil {
		return err
	}
	return query.Generate(plugin, files)
}
