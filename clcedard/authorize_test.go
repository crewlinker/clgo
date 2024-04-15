package clcedard_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	_ "embed"

	"github.com/crewlinker/clgo/clcedard"
	"github.com/crewlinker/clgo/clzap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
)

//go:embed testdata/ats/policies.cedar
var policies1 string

//go:embed testdata/ats/cedarschema.json
var schema1 []byte

// some example entities.
var entities1 = []any{
	map[string]any{
		"uid": map[string]any{
			"type": "Ats::User",
			"id":   "Alice",
		},
		"attrs": map[string]any{
			"ownerOfOrganizations": []any{map[string]any{
				"type": "Ats::Organization",
				"id":   "crewlinker",
			}},
		},
		"parents": []any{},
	},
}

var _ = Describe("authorize", func() {
	var ccl *clcedard.Client

	BeforeEach(func(ctx context.Context) {
		app := fx.New(
			fx.Populate(&ccl),
			clcedard.Provide(),
			clzap.TestProvide(),
			fx.Supply(http.DefaultClient),
		)

		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should do di", func(ctx context.Context) {
		Expect(ccl).ToNot(BeNil())
	})

	Describe("batch authorize", func() {
		var input clcedard.BatchInput

		BeforeEach(func() {
			input = clcedard.BatchInput{
				Entities: entities1,
				Policies: policies1,
				Items: []clcedard.InputItem{
					{
						Principal: `Ats::User::"Alice"`,
						Action:    `Ats::Action::"initRealm"`,
						Resource:  `Ats::Application::"atsback"`,
						Context: map[string]any{
							"source_ip": "222.222.222.1",
						},
					},
				},
			}

			for i := range 9 {
				item := input.Items[0]
				item.Principal = fmt.Sprintf(`Ats::User::"user%d"`, i)
				input.Items = append(input.Items, item)
			}

			Expect(json.Unmarshal(schema1, &input.Schema)).To(Succeed())
		})

		It("empty batch", func(ctx context.Context) {
			input := clcedard.BatchInput{
				Policies: ``,
				Schema:   map[string]any{},
				Items:    []clcedard.InputItem{},
				Entities: []any{},
			}

			output, err := ccl.BatchAuthorize(ctx, &input)
			Expect(err).ToNot(HaveOccurred())
			Expect(output.Items).To(BeEmpty())
		})

		It("should do some actual checks", func(ctx context.Context) {
			output, err := ccl.BatchAuthorize(ctx, &input)
			Expect(err).ToNot(HaveOccurred())
			Expect(output.Items).To(HaveLen(10))
			Expect(output.Items[0].ErrorMessages).To(BeEmpty())
			Expect(output.Items[0].Decision).To(Equal("Allow"))
			Expect(output.Items[1].Decision).To(Equal("Deny"))
		})

		It("should do checks with BatchIsAuthorized", func(ctx context.Context) {
			res, err := ccl.BatchIsAuthorized(ctx, &input)
			Expect(err).To(MatchError(MatchRegexp(`does not exist`)))
			Expect(res).To(HaveLen(10))
			Expect(res[0]).To(BeTrue())
			Expect(res[1]).To(BeFalse())
		})
	})
})
