// Command protoc-gen-go-aip is a protoc plugin that emits two companion
// files for every input proto: _aip.pb.resource.go (resource-name parsers
// driven by google.api.resource / google.api.resource_reference) and
// _aip.pb.query.go (AIP-132/160/158 helpers driven by
// (protoc_contrib.aip.field_reference) on a request's filter /
// order_by fields).
//
// NOTE: this binary's name collides with go.einride.tech/aip's
// cmd/protoc-gen-go-aip. Install only one — they generate different APIs.
package main

import (
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/protoc-contrib/protoc-gen-go-aip/internal/generator"
)

func main() {
	opts := &generator.Options{}
	protogen.Options{
		ParamFunc: opts.Set,
	}.Run(func(plugin *protogen.Plugin) error {
		plugin.SupportedFeatures = uint64(
			pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL |
				pluginpb.CodeGeneratorResponse_FEATURE_SUPPORTS_EDITIONS,
		)
		plugin.SupportedEditionsMinimum = descriptorpb.Edition_EDITION_2023
		plugin.SupportedEditionsMaximum = descriptorpb.Edition_EDITION_2023
		return generator.Generate(plugin, opts)
	})
}
