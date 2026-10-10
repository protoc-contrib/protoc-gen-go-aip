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

For each `.proto` it emits up to two companion files in the same Go
package:

- **`*_aip.pb.resource.go`** — resource-name parsers and helpers driven
  by `google.api.resource` and `google.api.resource_reference`.
- **`*_aip.pb.query.go`** — AIP-160 CEL filter helpers on List requests
  that carry a `filter` field. The resource is read off the List method's
  response, so nothing needs annotating.

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
- **Resource references** — every singular string field annotated with
  `google.api.resource_reference` (including cross-package references)
  gains a `Parse<Field>()` method on the owning message that delegates to
  the referent's parser, reading the field through its getter so
  explicit-presence fields (proto2 `optional`, edition 2023) work. A
  repeated reference — AIP-231 Batch Get's `names` — and `type: "*"` are
  skipped. A type declared in several packages (v1 and v2 of one API)
  resolves to the referring package's own declaration, else to the only
  package declaring it; several candidates and none local is an error.
  Set the plugin option `allow_unresolved_refs=true` to skip references
  whose target type isn't in the compilation unit, or is ambiguous.
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
- **ID helpers for goverter** — a single-pattern resource whose only
  variable segment is UUID-typed also gets two free functions,
  `Format<Resource>Name(id uuid.UUID) string` and
  `Parse<Resource>ID(s string) (uuid.UUID, error)`, which convert between
  a whole resource name and its ID. They are plain functions so goverter's
  `extend` directive can use them, which a method expression or a struct
  literal can't express.
- **AIP-133 create IDs** — that same `Create<Resource>Request` gains a
  `Parse<Resource>ID()` method returning the ID the caller proposed, or
  `uuid.Nil` when `<resource>_id` is empty: AIP-133 reads that as "the
  server assigns one", and which kind of UUID is the server's choice, so
  the accessor mints nothing —
  `if id == uuid.Nil { id, _ = uuid.NewV7() }` says which at the call site.
  Only for a single-pattern resource with a UUID-typed own ID, as in
  `protoc-gen-rust-aip`: a string ID has no validity rule the schema
  states, and a multi-pattern resource's create request does not say which
  pattern it creates under. It shares a name with the free function above
  but reads something different: the bare ID the request carries, not a
  resource name.

### Query pass

A request is a List request when all of these hold — the same rule
`protoc-gen-rust-aip` applies:

- a service **in the same `.proto` file** has a method, neither client- nor
  server-streaming, that takes it;
- that method's response has **exactly one** repeated message field (a map
  does not count), whose type is the resource;
- the request is a **top-level** message of that file — nested messages are
  not considered — with a singular `string filter` field;
- the resource has at least one field with a CEL type.

For each, the plugin emits:

- **`<Request>FilterEnv`** — a `*cel.Env` declaring every resource field
  that has a CEL type (strings, numbers, bools, enums as ints,
  `Timestamp`, `Duration`). Nested messages, repeated fields and maps are
  skipped. It has **no macros and no optional syntax**: `exists`, `has` and
  the like would expand to comprehensions and presence tests, and `a.?b` to
  an optional, none of which has a reading as a query — the same grammar
  `protoc-gen-rust-aip`'s environment accepts.
- **`ParseFilter()`** — compiles the `filter` expression against that
  environment and returns the checked `*cel.Ast`, or `(nil, nil)` when
  the filter is empty or blank. Pass the AST to a query layer such as
  [`pgxcel`](https://github.com/pgx-contrib/pgxcel), whose column map
  decides which fields a client may actually filter by.

Ordering (AIP-132) and page tokens (AIP-158) are not generated: both only
make sense against the column map, ordering and SQL that actually run,
which the `.proto` can't see, so they belong to the query layer.
`protoc-gen-rust-aip` generates nothing for them either.

## Example

Given this `books.proto`:

```proto
syntax = "proto3";

package books.v1;

import "google/api/resource.proto";
import "google/protobuf/timestamp.proto";

message Book {
  option (google.api.resource) = {
    type: "library.example.com/Book"
    pattern: "books/{book}"
  };

  string name = 1;
  string title = 2;
  string author = 3;
  google.protobuf.Timestamp create_time = 4;
}

message ListBooksRequest {
  int32 page_size = 1;
  string page_token = 2;
  string filter = 3;
}

message ListBooksResponse {
  repeated Book books = 1;
  string next_page_token = 2;
}

service Library {
  // ListBooksResponse's repeated Book makes this the List method for Book.
  rpc ListBooks(ListBooksRequest) returns (ListBooksResponse);
}
```

`buf generate` produces two companion files. The highlights:

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

**`books_aip.pb.query.go`** — CEL filter parser:

```go
// Declares name, title, author and create_time.
var ListBooksFilterEnv *cel.Env

func (x *ListBooksRequest) ParseFilter() (*cel.Ast, error)
```

Call sites stay terse:

```go
name, err := ParseBookName("books/foo")        // BookName{BookID: "foo"}, nil

filter, err := req.ParseFilter()               // *cel.Ast checked against ListBooksFilterEnv; nil when blank
where, args, err := pgxcel.Where(filter, pgxcel.WithColumns(columns))
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
| `allow_unresolved_refs` | `false` | When `true`, `google.api.resource_reference` fields whose target type is not in the compilation unit, or is declared only in several other packages, are skipped silently rather than producing a codegen error. |

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
