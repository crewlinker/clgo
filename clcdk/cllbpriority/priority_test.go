package cllbpriority_test

import (
	"testing"

	"github.com/crewlinker/clgo/clcdk/cllbpriority"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInfracon(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "priority")
}

var _ = Describe("priority", func() {
	It("should create priorities", func() {
		Expect(cllbpriority.ListenerPriority(0, 0, 0)).To(Equal(1000))
		Expect(cllbpriority.ListenerPriority(cllbpriority.MaxI, cllbpriority.MaxJ, cllbpriority.MaxK)).To(Equal(cllbpriority.MaxEncoded))
		Expect(cllbpriority.ListenerPriority(cllbpriority.MaxI+1, cllbpriority.MaxJ+1, cllbpriority.MaxK+1)).To(Equal(1000))
	})

	It("should not overlap", func() {
		prios := map[int]struct{}{}

		for i := range cllbpriority.MaxI {
			for j := range cllbpriority.MaxJ {
				for k := range cllbpriority.MaxK {
					prio := cllbpriority.ListenerPriority(i, j, k)

					Expect(prio).ToNot(BeKeyOf(prios))
					prios[prio] = struct{}{}
				}
			}
		}
	})
})
