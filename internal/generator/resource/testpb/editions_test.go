package testpb_test

import (
	googleuuid "github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"

	"github.com/protoc-contrib/protoc-gen-go-aip/internal/generator/resource/testpb/editions"
)

// That this file compiles is most of the test: in edition 2023 every field
// below is a *string, which the accessors read through their getters.
var _ = Describe("edition 2023 fields with explicit presence", func() {
	id := googleuuid.MustParse("55555555-5555-4555-8555-555555555555")

	It("parses an optional resource name and reference", func() {
		gadget := &editions.Gadget{
			Name: proto.String("gadgets/" + id.String()),
			Twin: proto.String("gadgets/" + id.String()),
		}
		name, err := gadget.ParseName()
		Expect(err).NotTo(HaveOccurred())
		Expect(name.GadgetID).To(Equal(id))
		twin, err := gadget.ParseTwin()
		Expect(err).NotTo(HaveOccurred())
		Expect(twin).To(Equal(name))
	})

	It("reads an unset name as empty", func() {
		_, err := (&editions.Gadget{}).ParseName()
		Expect(err).To(HaveOccurred())
	})

	It("reads an optional create ID", func() {
		got, err := (&editions.CreateGadgetRequest{}).ParseGadgetID()
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(googleuuid.Nil))

		got, err = (&editions.CreateGadgetRequest{GadgetId: proto.String(id.String())}).ParseGadgetID()
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(id))
	})
})
