package testpb_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	testpb "github.com/protoc-contrib/protoc-gen-go-aip/internal/generator/fieldmask/testpb"
)

var _ = Describe("Generated Validate()", func() {
	Describe("UpdateBookRequest.Validate", func() {
		It("accepts a nil mask (full replacement)", func() {
			req := &testpb.UpdateBookRequest{Book: &testpb.Book{}}
			Expect(req.Validate()).To(Succeed())
		})

		It("accepts paths that resolve to fields on Book", func() {
			req := &testpb.UpdateBookRequest{
				Book:       &testpb.Book{},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"title", "author"}},
			}
			Expect(req.Validate()).To(Succeed())
		})

		It(`accepts "*" as the sole path`, func() {
			req := &testpb.UpdateBookRequest{
				Book:       &testpb.Book{},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"*"}},
			}
			Expect(req.Validate()).To(Succeed())
		})

		It("rejects paths that do not exist on Book", func() {
			req := &testpb.UpdateBookRequest{
				Book:       &testpb.Book{},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"nonexistent"}},
			}
			Expect(req.Validate()).To(HaveOccurred())
		})

		It(`rejects "*" combined with other paths`, func() {
			req := &testpb.UpdateBookRequest{
				Book:       &testpb.Book{},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"*", "title"}},
			}
			Expect(req.Validate()).To(HaveOccurred())
		})
	})

	Describe("PatchBookRequest.Validate", func() {
		It("validates against the Target field even with non-conventional naming", func() {
			req := &testpb.PatchBookRequest{
				Target: &testpb.Book{},
				Mask:   &fieldmaskpb.FieldMask{Paths: []string{"read_count"}},
			}
			Expect(req.Validate()).To(Succeed())
		})

		It("rejects unknown paths against the Target field", func() {
			req := &testpb.PatchBookRequest{
				Target: &testpb.Book{},
				Mask:   &fieldmaskpb.FieldMask{Paths: []string{"isbn"}},
			}
			Expect(req.Validate()).To(HaveOccurred())
		})
	})
})
