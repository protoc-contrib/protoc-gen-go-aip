package testpb_test

import (
	"errors"
	stduuid "uuid"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/protoc-contrib/aip-go"
	"github.com/protoc-contrib/protoc-gen-go-aip/internal/generator/resource/testpb/uuid"
)

var _ = Describe("UUID-typed resource names", func() {
	It("parses a well-formed UUID4 segment into a typed struct", func() {
		id := stduuid.MustParse("11111111-1111-4111-8111-111111111111")

		got, err := uuid.ParseCollectionName("collections/" + id.String())
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(uuid.CollectionName{CollectionID: id}))
		Expect(got.String()).To(Equal("collections/" + id.String()))
	})

	It("wraps the underlying uuid.Parse error and names the segment index", func() {
		_, err := uuid.ParseCollectionName("collections/not-a-uuid")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(`parse "collections/not-a-uuid": segment 1:`))
	})

	// AIP-122 ids are lowercase. uuid.Parse also takes these spellings of the
	// same UUID, which would format back differently.
	DescribeTable("rejects a UUID segment that is not canonical",
		func(segment string) {
			_, err := uuid.ParseCollectionName("collections/" + segment)
			Expect(err).To(MatchError(ContainSubstring("segment 1: not a canonical lowercase UUID")))
		},
		Entry("uppercase", "11111111-1111-4111-8111-11111111111A"),
		Entry("braces", "{11111111-1111-4111-8111-111111111111}"),
		Entry("urn prefix", "urn:uuid:11111111-1111-4111-8111-111111111111"),
		Entry("no hyphens", "11111111111141118111111111111111"),
	)

	It("round-trips a UUID id through Format and Parse helpers", func() {
		id := stduuid.MustParse("44444444-4444-4444-8444-444444444444")

		name := uuid.FormatCollectionName(id)
		Expect(name).To(Equal("collections/" + id.String()))

		got, err := uuid.ParseCollectionID(name)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(id))
	})

	It("surfaces a parse error from ParseCollectionID when the name is malformed", func() {
		_, err := uuid.ParseCollectionID("collections/not-a-uuid")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(`parse "collections/not-a-uuid": segment 1:`))
	})

	Describe("the AIP-133 create-ID accessor", func() {
		It("is uuid.Nil when the caller proposes no ID", func() {
			got, err := (&uuid.CreateCollectionRequest{}).ParseCollectionID()
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(stduuid.Nil()))
		})

		It("parses the bare ID the caller proposed", func() {
			id := stduuid.MustParse("44444444-4444-4444-8444-444444444444")

			got, err := (&uuid.CreateCollectionRequest{CollectionId: id.String()}).ParseCollectionID()
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(id))
			Expect(uuid.CollectionName{CollectionID: got}.String()).To(Equal("collections/" + id.String()))
		})

		It("names the field when the proposed ID is not a UUID", func() {
			_, err := (&uuid.CreateCollectionRequest{CollectionId: "not-a-uuid"}).ParseCollectionID()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`parse collection_id "not-a-uuid":`))

			var field *aip.FieldError
			Expect(errors.As(err, &field)).To(BeTrue(), "%v is not a *aip.FieldError", err)
			Expect(field.Field()).To(Equal("collection_id"))
		})

		It("refuses a proposed ID that is not a canonical lowercase UUID", func() {
			_, err := (&uuid.CreateCollectionRequest{CollectionId: "11111111-1111-4111-8111-11111111111A"}).ParseCollectionID()
			Expect(err).To(MatchError(`collection_id: parse collection_id "11111111-1111-4111-8111-11111111111A": not a canonical lowercase UUID`))
		})

		It("refuses the nil UUID, which would read as no ID at all", func() {
			_, err := (&uuid.CreateCollectionRequest{CollectionId: stduuid.Nil().String()}).ParseCollectionID()
			Expect(err).To(MatchError(ContainSubstring("the nil UUID is not an ID")))
		})

		It("reads a nested resource's own ID, not its parent's", func() {
			id := stduuid.MustParse("55555555-5555-4555-8555-555555555555")

			got, err := (&uuid.CreateItemRequest{ItemId: id.String()}).ParseItemID()
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(id))
		})
	})

	It("flows typed parent ids through Parent() without string bridging", func() {
		orgID := stduuid.MustParse("22222222-2222-4222-8222-222222222222")
		itemID := stduuid.MustParse("33333333-3333-4333-8333-333333333333")

		child := uuid.ItemName{OrganizationID: orgID, ItemID: itemID}
		Expect(child.Parent()).To(Equal(uuid.OrganizationName{OrganizationID: orgID}))
		Expect(child.String()).To(Equal("organizations/" + orgID.String() + "/items/" + itemID.String()))
	})
})
