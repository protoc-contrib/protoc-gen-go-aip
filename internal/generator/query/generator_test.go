package query_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.einride.tech/aip/pagination"

	"github.com/protoc-contrib/protoc-gen-go-aip/internal/generator/query/testpb"
)

var _ = Describe("Generated AIP helpers", func() {
	Describe("{Resource}FilterDeclarations", func() {
		It("is initialised at package load", func() {
			Expect(testpb.BookFilterDeclarations).NotTo(BeNil())
		})
	})

	Describe("{Resource}Columns", func() {
		It("maps proto field names to DB columns for fields with the column annotation", func() {
			Expect(testpb.BookColumns).To(Equal(map[string]string{
				"create_time": "created_at",
			}))
		})
	})

	Describe("{Resource}OrderByFields", func() {
		It("lists every orderable path in proto declaration order", func() {
			Expect(testpb.BookOrderByFields).To(Equal([]string{
				"title",
				"create_time",
			}))
		})
	})

	Describe("ListBooksRequest.ParseOrderBy", func() {
		It("parses empty order_by successfully", func() {
			order, err := (&testpb.ListBooksRequest{}).ParseOrderBy()
			Expect(err).NotTo(HaveOccurred())
			Expect(order).NotTo(BeNil())
			Expect(order.Fields).To(BeEmpty())
		})

		It("parses a single orderable field", func() {
			order, err := (&testpb.ListBooksRequest{OrderBy: "title"}).ParseOrderBy()
			Expect(err).NotTo(HaveOccurred())
			Expect(order.Fields).To(HaveLen(1))
			Expect(order.Fields[0].Path).To(Equal("title"))
			Expect(order.Fields[0].Desc).To(BeFalse())
		})

		It("parses multiple fields with descending sort", func() {
			order, err := (&testpb.ListBooksRequest{OrderBy: "create_time desc, title"}).ParseOrderBy()
			Expect(err).NotTo(HaveOccurred())
			Expect(order.Fields).To(HaveLen(2))
			Expect(order.Fields[0].Path).To(Equal("create_time"))
			Expect(order.Fields[0].Desc).To(BeTrue())
			Expect(order.Fields[1].Path).To(Equal("title"))
			Expect(order.Fields[1].Desc).To(BeFalse())
		})

		It("returns an error prefixed with invalid order_by when a path is not in the allow-list", func() {
			_, err := (&testpb.ListBooksRequest{OrderBy: "author"}).ParseOrderBy()
			Expect(err).To(MatchError(ContainSubstring("invalid order_by")))
		})

		It("returns an error prefixed with invalid order_by on a syntactically invalid order_by", func() {
			_, err := (&testpb.ListBooksRequest{OrderBy: "title bogus_direction"}).ParseOrderBy()
			Expect(err).To(MatchError(ContainSubstring("invalid order_by")))
		})
	})

	Describe("ListBooksRequest.ParseFilter", func() {
		It("parses empty filter successfully", func() {
			filter, err := (&testpb.ListBooksRequest{}).ParseFilter()
			Expect(err).NotTo(HaveOccurred())
			Expect(filter).NotTo(BeNil())
		})

		It("parses a string equality filter", func() {
			filter, err := (&testpb.ListBooksRequest{Filter: `title = "The Go Programming Language"`}).ParseFilter()
			Expect(err).NotTo(HaveOccurred())
			Expect(filter.CheckedExpr).NotTo(BeNil())
		})

		It("parses an int comparison filter", func() {
			filter, err := (&testpb.ListBooksRequest{Filter: "read_count > 100"}).ParseFilter()
			Expect(err).NotTo(HaveOccurred())
			Expect(filter.CheckedExpr).NotTo(BeNil())
		})

		It("parses a compound filter with AND", func() {
			filter, err := (&testpb.ListBooksRequest{Filter: `author = "Kernighan" AND read_count > 10`}).ParseFilter()
			Expect(err).NotTo(HaveOccurred())
			Expect(filter.CheckedExpr).NotTo(BeNil())
		})

		It("returns an error prefixed with invalid filter when referencing an undeclared ident", func() {
			_, err := (&testpb.ListBooksRequest{Filter: `isbn = "9780134190440"`}).ParseFilter()
			Expect(err).To(MatchError(ContainSubstring("invalid filter")))
		})

		It("returns an error prefixed with invalid filter on a type mismatch", func() {
			_, err := (&testpb.ListBooksRequest{Filter: `read_count = "many"`}).ParseFilter()
			Expect(err).To(MatchError(ContainSubstring("invalid filter")))
		})

		It("returns an error prefixed with invalid filter on a syntactically invalid filter", func() {
			_, err := (&testpb.ListBooksRequest{Filter: "title ="}).ParseFilter()
			Expect(err).To(MatchError(ContainSubstring("invalid filter")))
		})
	})

	Describe("ListBooksRequest.ParsePageToken", func() {
		It("returns a zero-offset token with a non-zero checksum on the first page", func() {
			pt, err := (&testpb.ListBooksRequest{Filter: `title = "x"`}).ParsePageToken()
			Expect(err).NotTo(HaveOccurred())
			Expect(pt.Offset).To(BeZero())
			Expect(pt.RequestChecksum).NotTo(BeZero())
		})

		It("round-trips an offset page token via PageToken.Next.String", func() {
			req := &testpb.ListBooksRequest{Filter: `title = "x"`, PageSize: 50}
			seed, err := req.ParsePageToken()
			Expect(err).NotTo(HaveOccurred())

			req.PageToken = seed.Next(req).String()
			got, err := req.ParsePageToken()
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Offset).To(Equal(int64(50)))
		})

		It("errors when the filter changes between pages", func() {
			seed, err := (&testpb.ListBooksRequest{Filter: `title = "a"`, PageSize: 10}).ParsePageToken()
			Expect(err).NotTo(HaveOccurred())
			tok := seed.Next(&testpb.ListBooksRequest{PageSize: 10}).String()
			_, err = (&testpb.ListBooksRequest{Filter: `title = "b"`, PageToken: tok}).ParsePageToken()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ListBooksRequest.ParseQuery", func() {
		It("bundles filter, order_by, and page_token for a first-page request", func() {
			q, err := (&testpb.ListBooksRequest{
				Filter:  `title = "x"`,
				OrderBy: "title",
			}).ParseQuery()
			Expect(err).NotTo(HaveOccurred())
			Expect(q.Filter.CheckedExpr).NotTo(BeNil())
			Expect(q.OrderBy.Fields).To(HaveLen(1))
			Expect(q.PageToken).To(BeAssignableToTypeOf(pagination.PageToken{}))
			Expect(q.PageToken.Offset).To(BeZero())
			Expect(q.PageToken.RequestChecksum).NotTo(BeZero())
		})

		It("propagates the page-token error when the token checksum is stale", func() {
			seed, err := (&testpb.ListBooksRequest{Filter: `title = "a"`, PageSize: 10}).ParsePageToken()
			Expect(err).NotTo(HaveOccurred())
			tok := seed.Next(&testpb.ListBooksRequest{PageSize: 10}).String()
			_, err = (&testpb.ListBooksRequest{Filter: `title = "b"`, PageToken: tok}).ParseQuery()
			Expect(err).To(HaveOccurred())
		})
	})
})
