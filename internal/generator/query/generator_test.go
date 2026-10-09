package query_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/protoc-contrib/protoc-gen-go-aip/internal/generator/query/testpb"
)

var _ = Describe("Generated AIP helpers", func() {
	Describe("{Request}FilterEnv", func() {
		It("is initialised at package load", func() {
			Expect(testpb.ListBooksFilterEnv).NotTo(BeNil())
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
})
