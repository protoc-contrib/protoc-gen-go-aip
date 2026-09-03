module github.com/protoc-contrib/protoc-gen-go-aip

go 1.25.8

require (
	github.com/google/cel-go v0.28.0
	github.com/google/uuid v1.6.0
	github.com/iancoleman/strcase v0.3.0
	github.com/onsi/ginkgo/v2 v2.32.1
	github.com/onsi/gomega v1.43.0
	github.com/protoc-contrib/aip-go v0.0.0-20260903102901-ba65222135d0
	google.golang.org/genproto/googleapis/api v0.0.0-20260831171406-18b4a7587f8a
	google.golang.org/protobuf v1.36.12
)

require (
	cel.dev/expr v0.25.1 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20260402051712-545e8a4df936 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/exp v0.0.0-20240823005443-9b4947da3948 // indirect
	golang.org/x/mod v0.36.0 // indirect
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/sync v0.21.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
	golang.org/x/text v0.38.0 // indirect
	golang.org/x/tools v0.45.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260825221802-da73d73af1c5 // indirect
)

tool (
	github.com/onsi/ginkgo/v2/ginkgo
	google.golang.org/protobuf/cmd/protoc-gen-go
)
