package testpb_test

import (
	googleuuid "github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/protoc-contrib/protoc-gen-go-aip/internal/generator/resource/testpb/uuid"
)

var _ = Describe("UUID-typed resource names", func() {
	It("parses a well-formed UUID4 segment into a typed struct", func() {
		id := googleuuid.MustParse("11111111-1111-4111-8111-111111111111")

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

	Describe("the AIP-133 create-ID accessor", func() {
		It("is uuid.Nil when the caller proposes no ID", func() {
			got, err := (&uuid.CreateCollectionRequest{}).ParseCollectionID()
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(googleuuid.Nil))
		})

		It("parses the bare ID the caller proposed", func() {
			id := googleuuid.MustParse("44444444-4444-4444-8444-444444444444")

			got, err := (&uuid.CreateCollectionRequest{CollectionId: id.String()}).ParseCollectionID()
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(id))
			Expect(uuid.CollectionName{CollectionID: got}.String()).To(Equal("collections/" + id.String()))
		})

		It("names the field when the proposed ID is not a UUID", func() {
			_, err := (&uuid.CreateCollectionRequest{CollectionId: "not-a-uuid"}).ParseCollectionID()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(`parse collection_id "not-a-uuid":`))
		})

		It("refuses the nil UUID, which would read as no ID at all", func() {
			_, err := (&uuid.CreateCollectionRequest{CollectionId: googleuuid.Nil.String()}).ParseCollectionID()
			Expect(err).To(MatchError(ContainSubstring("the nil UUID is not an ID")))
		})

		It("reads a nested resource's own ID, not its parent's", func() {
			id := googleuuid.MustParse("55555555-5555-4555-8555-555555555555")

			got, err := (&uuid.CreateItemRequest{ItemId: id.String()}).ParseItemID()
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(id))
		})
	})

	It("flows typed parent ids through Parent() without string bridging", func() {
		orgID := googleuuid.MustParse("22222222-2222-4222-8222-222222222222")
		itemID := googleuuid.MustParse("33333333-3333-4333-8333-333333333333")

		child := uuid.ItemName{OrganizationID: orgID, ItemID: itemID}
		Expect(child.Parent()).To(Equal(uuid.OrganizationName{OrganizationID: orgID}))
		Expect(child.String()).To(Equal("organizations/" + orgID.String() + "/items/" + itemID.String()))
	})
})
