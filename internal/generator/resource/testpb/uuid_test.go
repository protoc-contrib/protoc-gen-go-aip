package testpb_test

import (
	"errors"
	stduuid "uuid"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/protoc-contrib/aip-go"
	"github.com/protoc-contrib/protoc-gen-go-aip/internal/generator/resource/testpb/uuid"
)

var _ = Describe("UUID-validated resource names", func() {
	It("parses a well-formed UUID4 segment, which the name keeps as a string", func() {
		id := stduuid.MustParse("11111111-1111-4111-8111-111111111111")

		got, err := uuid.ParseCollectionName("collections/" + id.String())
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(uuid.CollectionName{CollectionID: id.String()}))
		Expect(got.String()).To(Equal("collections/" + id.String()))
		Expect(got.Validate()).To(Succeed())
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

	Describe("the typed id", func() {
		It("is the own segment's UUID", func() {
			id := stduuid.MustParse("44444444-4444-4444-8444-444444444444")

			name, err := uuid.ParseCollectionName(uuid.FormatCollectionName(id))
			Expect(err).NotTo(HaveOccurred())
			got, err := name.ParseCollectionID()
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(id))
		})

		It("fails for a value that is not a UUID, which only a name built by hand holds", func() {
			_, err := uuid.CollectionName{CollectionID: "not-a-uuid"}.ParseCollectionID()
			Expect(err).To(MatchError(ContainSubstring("collection:")))
		})

		It("a parent's id is read through Parent()", func() {
			org := stduuid.MustParse("22222222-2222-4222-8222-222222222222")
			item := stduuid.MustParse("33333333-3333-4333-8333-333333333333")

			name, err := uuid.ParseItemName("organizations/" + org.String() + "/items/" + item.String())
			Expect(err).NotTo(HaveOccurred())
			got, err := name.Parent().ParseOrganizationID()
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(org))
			own, err := name.ParseItemID()
			Expect(err).NotTo(HaveOccurred())
			Expect(own).To(Equal(item))
		})
	})

	Describe("the AIP-159 wildcard", func() {
		It("is a plain `-` in a UUID segment, which the name keeps and prints", func() {
			got, err := uuid.ParseCollectionName("collections/-")
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(uuid.CollectionName{CollectionID: "-"}))
			Expect(got.String()).To(Equal("collections/-"))
			Expect(got.ContainsWildcard()).To(BeTrue())
		})

		It("is allowed for the parent of a child, which keeps its own id", func() {
			id := stduuid.MustParse("22222222-2222-4222-8222-222222222222")

			got, err := uuid.ParseItemName("organizations/-/items/" + id.String())
			Expect(err).NotTo(HaveOccurred())
			Expect(got.ContainsWildcard()).To(BeTrue())
			Expect(got.Parent().ContainsWildcard()).To(BeTrue())
			Expect(got.ItemID).To(Equal(id.String()))
			own, err := got.ParseItemID()
			Expect(err).NotTo(HaveOccurred())
			Expect(own).To(Equal(id))
		})

		It("is no id: the accessor refuses it, for the parent and for the leaf", func() {
			_, err := uuid.CollectionName{CollectionID: "-"}.ParseCollectionID()
			Expect(err).To(MatchError(ContainSubstring("the wildcard is not an id")))

			leaf, err := uuid.ParseItemName("organizations/22222222-2222-4222-8222-222222222222/items/-")
			Expect(err).NotTo(HaveOccurred())
			_, err = leaf.ParseItemID()
			Expect(err).To(MatchError(ContainSubstring("the wildcard is not an id")))
		})

		It("is not a valid name to write: Validate refuses it", func() {
			Expect(uuid.CollectionName{CollectionID: "-"}.Validate()).To(MatchError(ContainSubstring("not a canonical non-nil UUID")))
		})

		It("is not in a name without it", func() {
			got, err := uuid.ParseCollectionName("collections/11111111-1111-4111-8111-111111111111")
			Expect(err).NotTo(HaveOccurred())
			Expect(got.ContainsWildcard()).To(BeFalse())
		})
	})

	It("Validate refuses the nil UUID and a value that is not a UUID", func() {
		Expect(uuid.CollectionName{CollectionID: stduuid.Nil().String()}.Validate()).To(MatchError(ContainSubstring("not a canonical non-nil UUID")))
		Expect(uuid.CollectionName{CollectionID: "abc"}.Validate()).To(MatchError(ContainSubstring("not a canonical non-nil UUID")))
	})

	Describe("the free helpers", func() {
		It("round-trips an id through Format and Parse", func() {
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

		It("refuses a wildcard: it is not an id", func() {
			_, err := uuid.ParseCollectionID("collections/-")
			Expect(err).To(MatchError(ContainSubstring("the wildcard is not an id")))
		})

		It("reads a nested resource's own id and its parent's from the name", func() {
			org := stduuid.MustParse("22222222-2222-4222-8222-222222222222")
			item := stduuid.MustParse("33333333-3333-4333-8333-333333333333")
			name := "organizations/" + org.String() + "/items/" + item.String()

			own, err := uuid.ParseItemID(name)
			Expect(err).NotTo(HaveOccurred())
			Expect(own).To(Equal(item))
			parent, err := uuid.ParseItemOrganizationID(name)
			Expect(err).NotTo(HaveOccurred())
			Expect(parent).To(Equal(org))
		})

		It("refuses a wildcard parent: ParseItemOrganizationID is for a name that is concrete", func() {
			_, err := uuid.ParseItemOrganizationID("organizations/-/items/33333333-3333-4333-8333-333333333333")
			Expect(err).To(MatchError(ContainSubstring("the wildcard is not an id")))
		})
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
			Expect(uuid.FormatCollectionName(got)).To(Equal("collections/" + id.String()))
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

	It("builds a child from its parent and a typed id, and goes back up through Parent()", func() {
		orgID := stduuid.MustParse("22222222-2222-4222-8222-222222222222")
		itemID := stduuid.MustParse("33333333-3333-4333-8333-333333333333")

		child := uuid.OrganizationName{OrganizationID: orgID.String()}.ItemName(itemID)
		Expect(child).To(Equal(uuid.ItemName{OrganizationID: orgID.String(), ItemID: itemID.String()}))
		Expect(child.Parent()).To(Equal(uuid.OrganizationName{OrganizationID: orgID.String()}))
		Expect(child.String()).To(Equal("organizations/" + orgID.String() + "/items/" + itemID.String()))
	})
})
