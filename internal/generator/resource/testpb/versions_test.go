package testpb_test

import (
	googleuuid "github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "github.com/protoc-contrib/protoc-gen-go-aip/internal/generator/resource/testpb/versions/v1"
	v2 "github.com/protoc-contrib/protoc-gen-go-aip/internal/generator/resource/testpb/versions/v2"
)

// v1 and v2 declare the same resource types; only v2 types Shelf's ID as a
// UUID. The typed assignments below would not compile if either version's
// references or parents bound to the other's declarations.
var _ = Describe("one resource type declared in two API versions", func() {
	It("binds v1 to v1's string-typed Shelf", func() {
		var shelf v1.ShelfName
		var err error
		shelf, err = (&v1.Book{Shelf: "shelves/s1"}).ParseShelf()
		Expect(err).NotTo(HaveOccurred())
		Expect(shelf.ShelfID).To(Equal("s1"))

		book := v1.BookName{ShelfID: "s1", BookID: "b1"}
		shelf = book.Parent()
		Expect(shelf.String()).To(Equal("shelves/s1"))
	})

	It("binds v2 to v2's UUID-typed Shelf", func() {
		id := googleuuid.MustParse("66666666-6666-4666-8666-666666666666")
		var shelf v2.ShelfName
		var err error
		shelf, err = (&v2.Book{Shelf: "shelves/" + id.String()}).ParseShelf()
		Expect(err).NotTo(HaveOccurred())
		Expect(shelf.ShelfID).To(Equal(id))

		book := v2.BookName{ShelfID: id, BookID: "b1"}
		shelf = book.Parent()
		Expect(shelf.ShelfID).To(Equal(id))
	})
})
