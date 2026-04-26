package query_test

import (
	"os"
	"strings"
	"testing"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/protoc-contrib/protoc-gen-go-aip/internal/generator/query"
	_ "github.com/protoc-contrib/protoc-gen-go-aip/internal/generator/query/testpb"
	aippb "github.com/protoc-contrib/protoc-gen-go-aip/protoc_contrib/aip"
)

const (
	testProtoPath   = "internal/generator/query/testpb/test.proto"
	testGoldenPath  = "testpb/test_aip.pb.query.go"
	testGoldenWant  = "internal/generator/query/testpb/test_aip.pb.query.go"
	moduleParameter = "module=github.com/protoc-contrib/protoc-gen-go-aip"
)

func TestGenerate_Golden(t *testing.T) {
	plugin := newPlugin(t, buildRequest(t, []string{testProtoPath}, allRegisteredFiles()))
	if err := query.Generate(plugin); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	resp := plugin.Response()
	if resp.Error != nil {
		t.Fatalf("plugin reported error: %s", *resp.Error)
	}
	if len(resp.File) != 1 {
		t.Fatalf("want 1 generated file, got %d", len(resp.File))
	}
	if got := resp.File[0].GetName(); got != testGoldenWant {
		t.Fatalf("generated filename = %q, want %q", got, testGoldenWant)
	}

	want, err := os.ReadFile(testGoldenPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if got := resp.File[0].GetContent(); got != string(want) {
		t.Errorf("generated content does not match committed golden\n--- got:\n%s\n--- want:\n%s", got, want)
	}
}

func TestGenerate_SkipsFilesWithoutFieldReference(t *testing.T) {
	fd := syntheticFileDescriptor(t, "plain.proto", func(fd *descriptorpb.FileDescriptorProto) {
		// A resource with a List request/response pair but no field_reference
		// on the request fields: generator should emit nothing.
		fd.MessageType = []*descriptorpb.DescriptorProto{
			resourceMessage("Widget", "example/Widget", scalarField("name", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING)),
			plainListRequest("ListWidgetsRequest"),
			listResponseMessage("ListWidgetsResponse", ".plain.Widget"),
		}
	})
	plugin := newPlugin(t, buildRequest(t, []string{"plain.proto"}, append(allRegisteredFiles(), fd)))
	if err := query.Generate(plugin); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if n := len(plugin.Response().File); n != 0 {
		t.Fatalf("expected no generated files, got %d", n)
	}
}

func TestGenerate_RejectsUnsupportedFilterableKind(t *testing.T) {
	// field_reference exposes a repeated string field on the resource — celTypeExpr rejects it.
	fd := syntheticFileDescriptor(t, "bad_repeated.proto", func(fd *descriptorpb.FileDescriptorProto) {
		fd.MessageType = []*descriptorpb.DescriptorProto{
			resourceMessage("Widget", "example/Widget",
				repeatedField("tags", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			),
			fieldReferenceListRequest("ListWidgetsRequest", "example/Widget", []string{"tags"}, nil),
			listResponseMessage("ListWidgetsResponse", ".bad_repeated.Widget"),
		}
	})
	plugin := newPlugin(t, buildRequest(t, []string{"bad_repeated.proto"}, append(allRegisteredFiles(), fd)))
	err := query.Generate(plugin)
	if err == nil {
		t.Fatalf("expected error for repeated filterable field, got nil")
	}
	if !strings.Contains(err.Error(), "repeated") {
		t.Errorf("error = %q, want mention of repeated", err)
	}
}

func TestGenerate_RejectsUnsupportedMessageKind(t *testing.T) {
	// field_reference exposes a nested non-WKT message field — celTypeExpr rejects it.
	fd := syntheticFileDescriptor(t, "bad_msg.proto", func(fd *descriptorpb.FileDescriptorProto) {
		fd.MessageType = []*descriptorpb.DescriptorProto{
			{Name: proto.String("Inner")},
			resourceMessage("Widget", "example/Widget",
				messageField("inner", 1, ".bad_msg.Inner"),
			),
			fieldReferenceListRequest("ListWidgetsRequest", "example/Widget", []string{"inner"}, nil),
			listResponseMessage("ListWidgetsResponse", ".bad_msg.Widget"),
		}
	})
	plugin := newPlugin(t, buildRequest(t, []string{"bad_msg.proto"}, append(allRegisteredFiles(), fd)))
	err := query.Generate(plugin)
	if err == nil {
		t.Fatalf("expected error for nested message filterable field, got nil")
	}
	if !strings.Contains(err.Error(), "nested message") {
		t.Errorf("error = %q, want mention of nested message", err)
	}
}

func TestGenerate_ErrorsOnUnknownReferenceType(t *testing.T) {
	// field_reference names a resource type that no message in the compilation unit declares.
	fd := syntheticFileDescriptor(t, "unknown_ref.proto", func(fd *descriptorpb.FileDescriptorProto) {
		fd.MessageType = []*descriptorpb.DescriptorProto{
			fieldReferenceListRequest("ListWidgetsRequest", "example/Missing", []string{"name"}, nil),
		}
	})
	plugin := newPlugin(t, buildRequest(t, []string{"unknown_ref.proto"}, append(allRegisteredFiles(), fd)))
	err := query.Generate(plugin)
	if err == nil {
		t.Fatalf("expected error for unresolved type, got nil")
	}
	if !strings.Contains(err.Error(), "type") {
		t.Errorf("error = %q, want mention of type", err)
	}
}

// --- helpers ---

func newPlugin(t *testing.T, req *pluginpb.CodeGeneratorRequest) *protogen.Plugin {
	t.Helper()
	plugin, err := protogen.Options{}.New(req)
	if err != nil {
		t.Fatalf("protogen.Options.New: %v", err)
	}
	return plugin
}

func buildRequest(t *testing.T, toGenerate []string, files []*descriptorpb.FileDescriptorProto) *pluginpb.CodeGeneratorRequest {
	t.Helper()
	return &pluginpb.CodeGeneratorRequest{
		FileToGenerate: toGenerate,
		ProtoFile:      files,
		Parameter:      proto.String(moduleParameter),
	}
}

func allRegisteredFiles() []*descriptorpb.FileDescriptorProto {
	var out []*descriptorpb.FileDescriptorProto
	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		out = append(out, protodesc.ToFileDescriptorProto(fd))
		return true
	})
	return out
}

func syntheticFileDescriptor(t *testing.T, path string, build func(*descriptorpb.FileDescriptorProto)) *descriptorpb.FileDescriptorProto {
	t.Helper()
	pkg := strings.TrimSuffix(path, ".proto")
	fd := &descriptorpb.FileDescriptorProto{
		Name:    proto.String(path),
		Package: proto.String(pkg),
		Syntax:  proto.String("proto3"),
		Dependency: []string{
			"google/api/resource.proto",
			"protoc_contrib/aip/query.proto",
		},
		Options: &descriptorpb.FileOptions{
			GoPackage: proto.String("example.com/" + pkg + ";" + pkg),
		},
	}
	build(fd)
	return fd
}

func resourceMessage(name, resourceType string, fields ...*descriptorpb.FieldDescriptorProto) *descriptorpb.DescriptorProto {
	opts := &descriptorpb.MessageOptions{}
	proto.SetExtension(opts, annotations.E_Resource, &annotations.ResourceDescriptor{
		Type:    resourceType,
		Pattern: []string{strings.ToLower(name) + "s/{" + strings.ToLower(name) + "}"},
	})
	return &descriptorpb.DescriptorProto{
		Name:    proto.String(name),
		Field:   fields,
		Options: opts,
	}
}

func plainListRequest(name string) *descriptorpb.DescriptorProto {
	return &descriptorpb.DescriptorProto{
		Name: proto.String(name),
		Field: []*descriptorpb.FieldDescriptorProto{
			scalarField("filter", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			scalarField("order_by", 2, descriptorpb.FieldDescriptorProto_TYPE_STRING),
		},
	}
}

func fieldReferenceListRequest(name, refType string, filterFields, orderByFields []string) *descriptorpb.DescriptorProto {
	fields := []*descriptorpb.FieldDescriptorProto{
		fieldReferenceField("filter", 1, refType, filterFields),
	}
	if len(orderByFields) > 0 {
		fields = append(fields, fieldReferenceField("order_by", 2, refType, orderByFields))
	}
	return &descriptorpb.DescriptorProto{
		Name:  proto.String(name),
		Field: fields,
	}
}

func listResponseMessage(name, repeatedType string) *descriptorpb.DescriptorProto {
	return &descriptorpb.DescriptorProto{
		Name: proto.String(name),
		Field: []*descriptorpb.FieldDescriptorProto{
			{
				Name:     proto.String("items"),
				Number:   proto.Int32(1),
				Label:    descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
				Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
				TypeName: proto.String(repeatedType),
			},
		},
	}
}

func fieldReferenceField(name string, number int32, refType string, refFields []string) *descriptorpb.FieldDescriptorProto {
	opts := &descriptorpb.FieldOptions{}
	proto.SetExtension(opts, aippb.E_FieldReference, &aippb.FieldReference{
		Type:   refType,
		Fields: refFields,
	})
	return &descriptorpb.FieldDescriptorProto{
		Name:    proto.String(name),
		Number:  proto.Int32(number),
		Label:   descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
		Type:    descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum(),
		Options: opts,
	}
}

func scalarField(name string, number int32, kind descriptorpb.FieldDescriptorProto_Type) *descriptorpb.FieldDescriptorProto {
	return &descriptorpb.FieldDescriptorProto{
		Name:   proto.String(name),
		Number: proto.Int32(number),
		Label:  descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
		Type:   kind.Enum(),
	}
}

func repeatedField(name string, number int32, kind descriptorpb.FieldDescriptorProto_Type) *descriptorpb.FieldDescriptorProto {
	return &descriptorpb.FieldDescriptorProto{
		Name:   proto.String(name),
		Number: proto.Int32(number),
		Label:  descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
		Type:   kind.Enum(),
	}
}

func messageField(name string, number int32, typeName string) *descriptorpb.FieldDescriptorProto {
	return &descriptorpb.FieldDescriptorProto{
		Name:     proto.String(name),
		Number:   proto.Int32(number),
		Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
		Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
		TypeName: proto.String(typeName),
	}
}
