// Package generator orchestrates the resource-name and query passes of
// protoc-gen-go-aip. Each pass is invoked sequentially per file and emits
// its own _aip.pb.*.go companion in the same Go package.
package generator

import (
	"fmt"
	"strconv"

	"github.com/protoc-contrib/protoc-gen-go-aip/internal/generator/resource"
)

// Options carries top-level plugin parameters that flow into one or both
// generation passes. Resource-pass options live on Resource so the resource
// package's Generate signature stays unchanged.
type Options struct {
	Resource resource.Options
}

// Set applies a single `name=value` plugin parameter. The signature matches
// what protogen.Options.ParamFunc expects.
func (o *Options) Set(name, value string) error {
	switch name {
	case "allow_unresolved_refs":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid value for %q: %w", name, err)
		}
		o.Resource.AllowUnresolvedRefs = v
		return nil
	default:
		return fmt.Errorf("unknown plugin option %q", name)
	}
}
