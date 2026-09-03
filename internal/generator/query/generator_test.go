package query_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	aip "github.com/protoc-contrib/aip-go"

	"github.com/protoc-contrib/protoc-gen-go-aip/internal/generator/query/testpb"
)

var _ = Describe("Generated AIP helpers", func() {
	Describe("{Request}FilterEnv", func() {
		It("is initialised at package load", func() {
			Expect(testpb.ListBooksFilterEnv).NotTo(BeNil())
		})
	})

	Describe("{Request}OrderByFields", func() {
		It("lists every orderable field of the resource, in declaration order", func() {
			Expect(testpb.ListBooksOrderByFields).To(Equal([]string{
				"name",
				"title",
				"author",
				"read_count",
				"published",
				"genre",
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

		DescribeTable("rejects a path that is not orderable",
			func(orderBy string) {
				_, err := (&testpb.ListBooksRequest{OrderBy: orderBy}).ParseOrderBy()
				Expect(err).To(MatchError(ContainSubstring("invalid order_by")))
			},
			Entry("unknown field", "isbn"),
			// Fields with no CEL type are not declared orderable.
			Entry("nested message", "cover"),
			Entry("repeated field", "tags"),
		)

		It("returns an error prefixed with invalid order_by on a syntactically invalid order_by", func() {
			_, err := (&testpb.ListBooksRequest{OrderBy: "title bogus_direction"}).ParseOrderBy()
			Expect(err).To(MatchError(ContainSubstring("invalid order_by")))
		})
	})

	Describe("ListBooksRequest.ParseFilter", func() {
		It("returns a nil AST when no filter was provided", func() {
			ast, err := (&testpb.ListBooksRequest{}).ParseFilter()
			Expect(err).NotTo(HaveOccurred())
			Expect(ast).To(BeNil())
		})

		DescribeTable("compiles a valid CEL expression",
			func(filter string) {
				ast, err := (&testpb.ListBooksRequest{Filter: filter}).ParseFilter()
				Expect(err).NotTo(HaveOccurred())
				Expect(ast).NotTo(BeNil())
			},
			Entry("string equality", `title == "The Go Programming Language"`),
			Entry("int comparison", "read_count > 100"),
			Entry("bool", "published"),
			Entry("enum compares as an int", "genre == 1"),
			Entry("timestamp comparison", `create_time > timestamp("2024-01-01T00:00:00Z")`),
			Entry("conjunction", `author == "Kernighan" && read_count > 10`),
			Entry("disjunction", `author == "Kernighan" || author == "Ritchie"`),
			Entry("negation", `!published`),
			Entry("string function", `title.startsWith("The")`),
		)

		DescribeTable("rejects an expression the environment cannot compile",
			func(filter string) {
				_, err := (&testpb.ListBooksRequest{Filter: filter}).ParseFilter()
				Expect(err).To(MatchError(ContainSubstring("invalid filter")))
			},
			Entry("undeclared ident", `isbn == "9780134190440"`),
			// A field with no CEL type is not declared, so it is undeclared
			// to the compiler rather than a special case.
			Entry("nested message field", `cover == "x"`),
			Entry("repeated field", `tags == "x"`),
			Entry("type mismatch", `read_count == "many"`),
			Entry("syntax error", "read_count >"),
			// AIP-160 syntax is no longer accepted: this surface is CEL.
			Entry("AIP-160 equality", `title = "x"`),
			Entry("AIP-160 conjunction", `title == "x" AND published`),
			Entry("AIP-160 has", `title:"x"`),
		)
	})

	Describe("ListBooksRequest.ParsePageToken", func() {
		It("returns a zero-offset token with a non-zero checksum on the first page", func() {
			pt, err := (&testpb.ListBooksRequest{Filter: `title == "x"`}).ParsePageToken()
			Expect(err).NotTo(HaveOccurred())
			Expect(pt.Offset).To(BeZero())
			Expect(pt.RequestChecksum).NotTo(BeZero())
		})

		It("round-trips an offset page token via PageToken.NextOffset", func() {
			req := &testpb.ListBooksRequest{Filter: `title == "x"`, PageSize: 50}
			seed, err := req.ParsePageToken()
			Expect(err).NotTo(HaveOccurred())

			encoded, err := seed.NextOffset(req).Encode()
			Expect(err).NotTo(HaveOccurred())

			req.PageToken = encoded
			got, err := req.ParsePageToken()
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Offset).To(Equal(int64(50)))
		})

		It("errors when the filter changes between pages", func() {
			seed, err := (&testpb.ListBooksRequest{Filter: `title == "a"`, PageSize: 10}).ParsePageToken()
			Expect(err).NotTo(HaveOccurred())
			encoded, err := seed.NextOffset(&testpb.ListBooksRequest{PageSize: 10}).Encode()
			Expect(err).NotTo(HaveOccurred())

			_, err = (&testpb.ListBooksRequest{Filter: `title == "b"`, PageToken: encoded}).ParsePageToken()
			Expect(err).To(MatchError(aip.ErrChecksumMismatch))
		})
	})

	Describe("ListBooksRequest.ParseQuery", func() {
		It("bundles filter, order_by, and page_token for a first-page request", func() {
			q, err := (&testpb.ListBooksRequest{
				Filter:  `title == "x"`,
				OrderBy: "title",
			}).ParseQuery()
			Expect(err).NotTo(HaveOccurred())
			Expect(q.Filter).NotTo(BeNil())
			Expect(q.OrderBy.Fields).To(HaveLen(1))
			Expect(q.PageToken).To(BeAssignableToTypeOf(aip.PageToken{}))
			Expect(q.PageToken.Offset).To(BeZero())
			Expect(q.PageToken.RequestChecksum).NotTo(BeZero())
		})

		It("propagates the page-token error when the token checksum is stale", func() {
			seed, err := (&testpb.ListBooksRequest{Filter: `title == "a"`, PageSize: 10}).ParsePageToken()
			Expect(err).NotTo(HaveOccurred())
			encoded, err := seed.NextOffset(&testpb.ListBooksRequest{PageSize: 10}).Encode()
			Expect(err).NotTo(HaveOccurred())

			_, err = (&testpb.ListBooksRequest{Filter: `title == "b"`, PageToken: encoded}).ParseQuery()
			Expect(err).To(MatchError(aip.ErrChecksumMismatch))
		})
	})
})
