package generator

import (
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/protoc-contrib/protoc-gen-go-aip/internal/generator/fieldmask"
	"github.com/protoc-contrib/protoc-gen-go-aip/internal/generator/query"
	"github.com/protoc-contrib/protoc-gen-go-aip/internal/generator/resource"
)

// Generate runs all passes against the plugin request. The resource pass
// emits *_aip.pb.resource.go and walks every file (so cross-file refs
// resolve); the query pass emits *_aip.pb.query.go from List<Resource>Request
// messages annotated with (protoc_contrib.aip.filterable|orderable); the
// fieldmask pass emits *_aip.pb.fieldmask.go with Validate() on AIP-134
// update-request shaped messages. Each pass is a no-op for files that have
// nothing to emit.
func Generate(plugin *protogen.Plugin, opts *Options) error {
	if opts == nil {
		opts = &Options{}
	}
	if err := resource.Generate(plugin, &opts.Resource); err != nil {
		return err
	}
	if err := query.Generate(plugin); err != nil {
		return err
	}
	return fieldmask.Generate(plugin)
}
