package clserve_test

import (
	"github.com/crewlinker/clgo/clserve"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("reverser", func() {
	var rev *clserve.Reverser
	BeforeEach(func() {
		rev = clserve.NewReverser()
		Expect(rev.Named("homepage", "/{$}")).To(Equal(`/{$}`))
		Expect(rev.Named("blog_post", "/blog/{id}/{$}")).To(Equal(`/blog/{id}/{$}`))
	})

	It("should reverse", func() {
		Expect(rev.Reverse("homepage")).To(Equal(`/`))
	})

	It("should error if already exists", func() {
		_, err := rev.NamedPattern("homepage", "/")
		Expect(err).To(MatchError(MatchRegexp(`already exists`)))
	})

	It("should panic for Named error", func() {
		Expect(func() {
			rev.Named("bogus", "")
		}).To(PanicWith(MatchRegexp(`failed to parse`)))
	})

	It("should error if reversing unknown name", func() {
		_, err := rev.Reverse("bogus")
		Expect(err).To(MatchError(MatchRegexp(`no pattern named: "bogus"`)))
	})

	It("should error if url building fails", func() {
		_, err := rev.Reverse("blog_post")
		Expect(err).To(MatchError(MatchRegexp(`not enough values`)))
	})
})
