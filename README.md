# protoc-gen-go-aip

[![CI](https://github.com/protoc-contrib/protoc-gen-go-aip/actions/workflows/ci.yml/badge.svg)](https://github.com/protoc-contrib/protoc-gen-go-aip/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/protoc-contrib/protoc-gen-go-aip?include_prereleases)](https://github.com/protoc-contrib/protoc-gen-go-aip/releases)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE.md)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![protoc](https://img.shields.io/badge/protoc-compatible-blue)](https://protobuf.dev)

A [protoc](https://protobuf.dev) plugin that emits Go helpers for [Google
AIP](https://aip.dev) resource patterns and List-RPC query handling. It is
a unification of two earlier `protoc-contrib` plugins, with selected
additions adopted from [`go.einride.tech/aip`](https://github.com/einride/aip-go).

For each `.proto` it emits up to three companion files in the same Go
package:

- **`*_aip.pb.resource.go`** — resource-name parsers and helpers driven
  by `google.api.resource` and `google.api.resource_reference`.
- **`*_aip.pb.query.go`** — AIP-160 filter / AIP-132 ordering /
  AIP-158 pagination helpers driven by `(protoc_contrib.aip.filterable)`
  and `(protoc_contrib.aip.orderable)` field options.
- **`*_aip.pb.fieldmask.go`** — `Validate()` on AIP-134 update-request
  shaped messages, delegating to
  [`go.einride.tech/aip/fieldmask.Validate`](https://pkg.go.dev/go.einride.tech/aip/fieldmask#Validate).

> **⚠ Binary-name collision.** This plugin's binary is `protoc-gen-go-aip`,
> the same name used by the upstream einride plugin under
> [`go.einride.tech/aip/cmd/protoc-gen-go-aip`](https://pkg.go.dev/go.einride.tech/aip/cmd/protoc-gen-go-aip).
> The two generate **different** APIs and are not interchangeable — install
> only one. If you need einride's resource-name layout (with its
> `MarshalString` / `UnmarshalString` interface and string-only segments),
> use einride's plugin. If you need this plugin's UUID-typed segments,
> cross-package references, and the AIP query helpers, use this one.

## Features

### Resource-name pass

- **Single-pattern resources** — emits `type <Type>Name struct { ... }`
  with one field per `{variable}` segment, plus `Parse<Type>Name`,
  `ParseFull<Type>Name`, `String()`, `FullName()`, `MarshalText` /
  `UnmarshalText`.
- **Multi-pattern resources** — emits a sealed `<Type>Name` interface
  and one struct per pattern named after its parent
  (e.g. `PublisherBookName`, `AuthorBookName`), plus a polymorphic
  `Parse<Type>Name` that tries each pattern in declaration order.
- **`Parent()` navigation** — child resources get a `Parent()` method
  returning the matched parent's generated type. Each pattern of a
  multi-pattern resource returns its own parent type.
- **Parent constructors** — the parent struct gains a method named after
  the child type that builds the child by inheriting parent fields and
  taking only the child-only segments as arguments
  (e.g. `parent.ProjectThingName(thingID)`).
- **Resource references** — every field annotated with
  `google.api.resource_reference` (including cross-package references)
  gains a `Parse<Field>()` method on the owning message that delegates to
  the referent's parser. Set the plugin option
  `allow_unresolved_refs=true` to skip references whose target type
  isn't in the compilation unit.
- **`Validate()` / `Type()` / `Pattern()` / `ContainsWildcard()`** —
  every generated struct exposes these AIP-122/159 helpers (adopted from
  einride's plugin).
- **File-level resources** — `google.api.resource_definition` at file
  scope emits parsers even without a backing message.
- **UUID-typed segments** — declare `(google.api.field_info).format = UUID4`
  on the `<var>_id` field of a `Create<Resource>Request` (AIP-133) to
  have the generated struct field typed as `uuid.UUID` and validated at
  parse time. The parent UUID consistency check is automatic across the
  pattern tree.

### Query pass

- **`ParseFilter()`** — emits a `filtering.Filter` parser using
  `filtering.Declarations` derived from `(protoc_contrib.aip.filterable)`
  fields.
- **`ParseOrderBy()`** — emits an `ordering.OrderBy` parser keyed off
  `(protoc_contrib.aip.orderable)` fields.
- **`ParsePageToken()`** — emits a `pagination.PageToken` parser when
  the `List<Resource>Request` has `page_token` and `page_size`.
- **`ParseQuery()`** — emits a single helper that runs all three at once
  and returns a typed `Query` struct.

### Fieldmask pass

- **`Validate()`** — for any message that pairs exactly one
  `google.protobuf.FieldMask` field with exactly one other singular
  message-typed field (the AIP-134 update-request shape, e.g.
  `UpdateBookRequest { Book book = 1; FieldMask update_mask = 2; }`),
  emits a `Validate()` method that delegates to
  `fieldmask.Validate(mask, target)`. A nil mask is accepted as full
  replacement; `"*"` is accepted only as the sole path; every other path
  must resolve to a field on the target message. Detection is purely
  structural — the rule applies regardless of whether the request is
  named `Update*Request`, `Patch*Request`, or otherwise. Messages with
  zero or two-plus message-typed fields are silently skipped to avoid
  emitting half-validated code.

## Example

Given this `books.proto`:

```proto
syntax = "proto3";

package books.v1;

import "google/api/resource.proto";
import "google/protobuf/field_mask.proto";
import "google/protobuf/timestamp.proto";
import "protoc_contrib/aip/query.proto";

message Book {
  option (google.api.resource) = {
    type: "library.example.com/Book"
    pattern: "books/{book}"
  };

  string name = 1;
  string title = 2 [
    (protoc_contrib.aip.filterable) = true,
    (protoc_contrib.aip.orderable) = true
  ];
  string author = 3 [(protoc_contrib.aip.filterable) = true];
  google.protobuf.Timestamp create_time = 4 [(protoc_contrib.aip.orderable) = true];
}

message ListBooksRequest {
  int32 page_size = 1;
  string page_token = 2;
  string filter = 3;
  string order_by = 4;
}

message ListBooksResponse {
  repeated Book books = 1;
  string next_page_token = 2;
}

message UpdateBookRequest {
  Book book = 1;
  google.protobuf.FieldMask update_mask = 2;
}
```

`buf generate` produces three companion files. The highlights:

**`books_aip.pb.resource.go`** — resource-name parser:

```go
type BookName struct {
    BookID string
}

func ParseBookName(s string) (BookName, error)
func (n BookName)  String() string
func (n BookName)  Validate() error
func (x *Book)     ParseName() (BookName, error)
```

**`books_aip.pb.query.go`** — List-RPC parsers:

```go
var BookOrderByFields = []string{"title", "create_time"}

func (x *ListBooksRequest) ParseFilter()    (filtering.Filter, error)
func (x *ListBooksRequest) ParseOrderBy()   (ordering.OrderBy, error)
func (x *ListBooksRequest) ParsePageToken() (pagination.PageToken, error)
func (x *ListBooksRequest) ParseQuery()     (Query, error)
```

**`books_aip.pb.fieldmask.go`** — update-mask validator:

```go
// Validate delegates to fieldmask.Validate(x.UpdateMask, x.Book).
func (x *UpdateBookRequest) Validate() error
```

Call sites stay terse:

```go
name, err := ParseBookName("books/foo")        // BookName{BookID: "foo"}, nil

q, err := req.ParseQuery()                     // filter + order_by + page_token in one call
//   q.Filter.CheckedExpr   — CEL-validated against Book's filterable fields
//   q.OrderBy.Fields       — paths checked against BookOrderByFields
//   q.PageToken.Offset     — pagination cursor; checksum verifies request stability

if err := updateReq.Validate(); err != nil {   // update_mask references a non-Book field
    return status.Error(codes.InvalidArgument, err.Error())
}
```

## Installation

```bash
go install github.com/protoc-contrib/protoc-gen-go-aip/cmd/protoc-gen-go-aip@latest
```

## Usage

### With buf

Add the plugin to your `buf.gen.yaml`:

```yaml
version: v2
plugins:
  - local: protoc-gen-go-aip
    out: .
    opt:
      - module=github.com/your-org/your-module
```

Then run:

```bash
buf generate
```

### With protoc

```bash
protoc \
  --go-aip_out=. \
  --go-aip_opt=module=github.com/your-org/your-module \
  -I proto/ \
  proto/example.proto
```

## Options

| Option                  | Default | Effect                                                                                                                  |
| ----------------------- | ------- | ----------------------------------------------------------------------------------------------------------------------- |
| `allow_unresolved_refs` | `false` | When `true`, `google.api.resource_reference` fields whose target type is not in the compilation unit are skipped silently rather than producing a codegen error. |

## Migration

Replace prior `protoc-gen-go-aip-resource` and `protoc-gen-go-aip-query`
plugin entries in `buf.gen.yaml` with a single `protoc-gen-go-aip`
entry. Output suffixes are unchanged (`_aip.pb.resource.go`,
`_aip.pb.query.go`), so downstream import paths don't churn.

## Contributing

To set up a development environment with [Nix](https://nixos.org):

```bash
nix develop
go test ./...
```

Or, without Nix, ensure `go`, `protoc`, and `buf` are on your `PATH`.

## License

[MIT](LICENSE.md)
