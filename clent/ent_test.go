package clent_test

import (
	"testing"

	"entgo.io/ent/dialect"
	"github.com/crewlinker/clgo/clent"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestClent(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clent")
}

var _ = Describe("clid field", func() {
	It("should panic with invalid length", func() {
		Expect(func() {
			clent.Field.CLID("id", "a")
		}).To(PanicWith(MatchRegexp(`clid prefix not of length`)))
	})

	It("should have constructed correct field", func() {
		field := clent.Field.CLID("id2", "orga")

		Expect(field.Descriptor().Name).To(Equal(`id2`))
		Expect(field.Descriptor().SchemaType).To(Equal(map[string]string{
			dialect.Postgres: "varchar(31)",
		}))
	})
})
