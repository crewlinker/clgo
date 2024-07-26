package cloapi_test

import (
	"context"
	"testing"

	"github.com/crewlinker/clgo/cloapi"
	"github.com/getkin/kin-openapi/openapi3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	_ "embed"
)

func TestOApi(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "cloapi")
}

//go:embed testdata/schema.yml
var testSchema []byte

var _ = Describe("schema", func() {
	var doc *openapi3.T
	BeforeEach(func() {
		var err error
		doc, err = cloapi.LoadSchemaTmpl(testSchema)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should have loaded schema", func(ctx context.Context) {
		Expect(doc.Validate(ctx)).To(Succeed())
		Expect(FirstOperation(doc).Extensions).ToNot(HaveKey("x-amazon-apigateway-integration"))
	})

	Describe("with decorations", func() {
		BeforeEach(func() {
			Expect(cloapi.DecorateSchemaTmpl(doc)).To(Succeed())
		})

		It("should be decorated", func() {
			Expect(FirstOperation(doc).Extensions).To(HaveKey("x-amazon-apigateway-integration"))
		})

		It("should be executable", func() {
			src, err := doc.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())

			res, sum, err := cloapi.ExecuteSchemaTmpl(src, cloapi.SchemaDeployment{
				Title: "foobar",
			})

			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(ContainSubstring(`"title":"foobar"`))
			Expect(sum).ToNot(Equal([32]byte{}))
		})
	})
})

func FirstOperation(doc *openapi3.T) *openapi3.Operation {
	for _, path := range doc.Paths.InMatchingOrder() {
		item := doc.Paths.Find(path)
		for _, op := range item.Operations() {
			return op
		}
	}

	Fail("didn't find at least one operation in the OpenAPI Spec")

	return nil
}
