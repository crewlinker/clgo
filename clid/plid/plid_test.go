package plid_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/crewlinker/clgo/clid/plid"
	"github.com/oklog/ulid/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	_ fmt.Stringer     = plid.ID{}
	_ sql.Scanner      = &plid.ID{}
	_ json.Marshaler   = plid.ID{}
	_ json.Unmarshaler = &plid.ID{}

	// https://pkg.go.dev/github.com/99designs/gqlgen/graphql#Marshaler
	_ interface{ MarshalGQL(w io.Writer) } = plid.ID{}
	// https://pkg.go.dev/github.com/99designs/gqlgen/graphql#Marshaler
	_ interface{ UnmarshalGQL(v interface{}) error } = &plid.ID{}
)

func TestClid(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clid/plid")
}

var _ = Describe("prefix ulid", func() {
	var entr io.Reader
	var mst uint64

	BeforeEach(func() {
		entr = rand.New(rand.NewSource(1))
		mst = ulid.Timestamp(time.Unix(1701767130, 0))
	})

	It("should generate new plid", func() {
		id := plid.New("usr")
		Expect(id.String()).To(HavePrefix("usr-"))
		Expect(id.String()).To(HaveLen(30))
	})

	It("should stringer zero value", func() {
		var id plid.ID
		Expect(id.String()).To(Equal(`zzz-00000000000000000000000000`))
	})

	It("should generate and stringer", func() {
		id, err := plid.NewFromParts("usr", mst, entr)
		Expect(err).ToNot(HaveOccurred())
		Expect(id.String()).To(Equal(`usr-01HGWKKAWGABYZR1S1G9JMY5HZ`))
	})

	It("should scan", func() {
		var id plid.ID
		Expect(id.Scan("usr-01HGWKKAWGABYZR1S1G9JMY5HZ")).To(Succeed())

		Expect(id.String()).To(Equal("usr-01HGWKKAWGABYZR1S1G9JMY5HZ"))
	})

	It("should marshal gql", func() {
		id, err := plid.NewFromParts("usr", mst, entr)
		Expect(err).ToNot(HaveOccurred())

		var buf bytes.Buffer
		id.MarshalGQL(&buf)
		Expect(buf.String()).To(Equal(`"usr-01HGWKKAWGABYZR1S1G9JMY5HZ"`))
	})

	It("should unmarshal gql", func() {
		var id plid.ID
		Expect(id.UnmarshalGQL("usr-01HGWKKAWGABYZR1S1G9JMY5HZ")).To(Succeed())

		Expect(id.String()).To(Equal("usr-01HGWKKAWGABYZR1S1G9JMY5HZ"))
	})

	It("should marshal json", func() {
		data, err := json.Marshal(struct{ ID plid.ID }{})
		Expect(err).ToNot(HaveOccurred())
		Expect(data).To(MatchJSON(`{"ID":"zzz-00000000000000000000000000"}`))
	})

	It("should unmarshal json", func() {
		var v struct{ ID plid.ID }
		Expect(json.Unmarshal([]byte(`{"ID":"usr-01HGWKKAWGABYZR1S1G9JMY5HZ"}`), &v)).To(Succeed())
		Expect(v.ID.String()).To(Equal(`usr-01HGWKKAWGABYZR1S1G9JMY5HZ`))
	})

	Describe("errors", func() {
		It("should error on invalid scan type", func() {
			var id plid.ID
			Expect(id.Scan(1)).To(MatchError(`plid: failed to scan int: type not supported`))
		})

		It("should error on invalid scan format", func() {
			var id plid.ID
			Expect(id.Scan("")).To(MatchError(`plid: failed to scan string: missing separator '-'`))
		})

		It("should error on invalid ulid format", func() {
			var id plid.ID
			Expect(id.Scan("x-z")).To(MatchError(`plid: ulid: bad data size when unmarshaling`))
		})

		It("should error on unmarshal json", func() {
			var v struct{ ID plid.ID }
			Expect(json.Unmarshal([]byte(`{"ID":1}`), &v)).
				To(MatchError(`plid: failed to unmarshal as string: json: cannot unmarshal number into Go value of type string`))
		})
	})
})
