package clcdkcr_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/crewlinker/clgo/clcdk/clcdkcr"
	"github.com/go-playground/validator/v10"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestClcdkcr(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "clcdk/clcdkcr")
}

var _ = Describe("resource handling", func() {
	var val *validator.Validate
	var logs *zap.Logger
	var hadl1 clcdkcr.Handler[res1Input, res1Output]
	var hadl2 clcdkcr.Handler[res2Input, res2Output]
	var hadl3 clcdkcr.Handler[res3Input, res3Output]

	BeforeEach(func() {
		val = validator.New()
		hadl1, hadl2, hadl3 = hdl1{}, hdl2{}, hdl3{}
		zc, _ := observer.New(zapcore.DebugLevel)
		logs = zap.New(zc)
	})

	It("should leave handled to empty if no resource matched", func(ctx context.Context) {
		res, err := clcdkcr.Handle3(ctx, logs, val, cfn.Event{ResourceType: "Bogus"}, hadl1, hadl2, hadl3)
		Expect(err).To(MatchError(MatchRegexp(`unsupported resource type: Bogus`)))
		Expect(res.Handled).To(Equal(""))
		Expect(res.Visited).To(Equal([]string{"Custom::Hdl1", "Custom::Hdl2", "Custom::Hdl3"}))
	})

	It("should visit some if one is errors", func(ctx context.Context) {
		res, err := clcdkcr.Handle3(ctx, logs, val, cfn.Event{
			ResourceType: "Custom::Hdl2",
			RequestType:  "Bogus",
			ResourceProperties: map[string]any{
				"Foo": "foo",
			},
		}, hadl1, hadl2, hadl3)
		Expect(err).To(MatchError(MatchRegexp(`unsupported request type: Bogus`)))
		Expect(res.Handled).To(Equal("Custom::Hdl2"))
		Expect(res.Visited).To(Equal([]string{"Custom::Hdl1", "Custom::Hdl2", "Custom::Hdl3"}))
	})

	Describe("create", func() {
		It("should encode/decode", func(ctx context.Context) {
			res, err := clcdkcr.Handle3(ctx, logs, val, cfn.Event{
				ResourceType: "Custom::Hdl2",
				RequestType:  "Create",
				ResourceProperties: map[string]any{
					"Foo": "foo",
				},
			}, hadl1, hadl2, hadl3)

			Expect(err).ToNot(HaveOccurred())
			Expect(res.Handled).To(Equal("Custom::Hdl2"))
			Expect(res.Visited).To(Equal([]string{"Custom::Hdl1", "Custom::Hdl2", "Custom::Hdl3"}))
			Expect(res.Output.Data).To(Equal(map[string]any{"Bar": "foofoocreated"}))
		})

		It("should validate resource properties", func(ctx context.Context) {
			_, err := clcdkcr.Handle3(ctx, logs, val, cfn.Event{
				ResourceType: "Custom::Hdl2",
				RequestType:  "Create",
				ResourceProperties: map[string]any{
					"Foo": "_",
				},
			}, hadl1, hadl2, hadl3)
			Expect(err).To(MatchError(MatchRegexp(`\(new\) resource properties`)))
			Expect(err).To(MatchError(MatchRegexp(`Field validation for 'Foo'`)))
		})
	})

	Describe("delete", func() {
		It("should encode/decode", func(ctx context.Context) {
			res, err := clcdkcr.Handle3(ctx, logs, val, cfn.Event{
				PhysicalResourceID: "Some-resource-id",
				ResourceType:       "Custom::Hdl2",
				RequestType:        "Delete",
				ResourceProperties: map[string]any{
					"Foo": "foo",
				},
			}, hadl1, hadl2, hadl3)

			Expect(err).ToNot(HaveOccurred())
			Expect(res.Output.Data).To(Equal(map[string]any{"Bar": "foofoodeleted"}))
			Expect(res.Output.PhysicalResourceID).To(Equal("Some-resource-id"))
		})
	})

	Describe("update", func() {
		It("should encode/decode", func(ctx context.Context) {
			res, err := clcdkcr.Handle3(ctx, logs, val, cfn.Event{
				ResourceType: "Custom::Hdl2",
				RequestType:  "Update",
				ResourceProperties: map[string]any{
					"Foo": "foo",
				},
				OldResourceProperties: map[string]any{
					"Foo": "oldfoo",
				},
			}, hadl1, hadl2, hadl3)

			Expect(err).ToNot(HaveOccurred())
			Expect(res.Output.Data).To(Equal(map[string]any{"Bar": "foooldfooupdated"}))
		})

		It("should validate resource properties", func(ctx context.Context) {
			res, err := clcdkcr.Handle2(ctx, logs, val, cfn.Event{
				ResourceType: "Custom::Hdl2",
				RequestType:  "Update",
				ResourceProperties: map[string]any{
					"Foo": "foo",
				},
				OldResourceProperties: map[string]any{
					"Foo": "_",
				},
			}, hadl1, hadl2)

			Expect(res.Visited).To(Equal([]string{"Custom::Hdl1", "Custom::Hdl2"}))
			Expect(err).To(MatchError(MatchRegexp(`old input properties`)))
			Expect(err).To(MatchError(MatchRegexp(`Field validation for 'Foo'`)))
		})
	})

	It("should leave pass on errors from implementation", func(ctx context.Context) {
		res, err := clcdkcr.Handle1(ctx, logs, val, cfn.Event{ResourceType: "Custom::Hdl3", RequestType: "Create"}, hadl3)
		Expect(err).To(MatchError(MatchRegexp(`failed to handle Create: some error`)))
		Expect(res.Handled).To(Equal("Custom::Hdl3"))
		Expect(res.Visited).To(Equal([]string{"Custom::Hdl3"}))
	})

	It("should Handle4", func(ctx context.Context) {
		_, err := clcdkcr.Handle4(ctx, logs, val, cfn.Event{}, hadl1, hadl1, hadl1, hadl1)
		Expect(err).To(MatchError(MatchRegexp(`unsupported resource type: `)))
	})
	It("should Handle5", func(ctx context.Context) {
		_, err := clcdkcr.Handle5(ctx, logs, val, cfn.Event{}, hadl1, hadl1, hadl1, hadl1, hadl1)
		Expect(err).To(MatchError(MatchRegexp(`unsupported resource type: `)))
	})
	It("should Handle6", func(ctx context.Context) {
		_, err := clcdkcr.Handle6(ctx, logs, val, cfn.Event{}, hadl1, hadl1, hadl1, hadl1, hadl1, hadl1)
		Expect(err).To(MatchError(MatchRegexp(`unsupported resource type: `)))
	})
})

// hdl1.
type hdl1 struct{}

type (
	res1Input  struct{}
	res1Output struct{}
)

func (h hdl1) Type() string { return "Custom::Hdl1" }
func (h hdl1) Create(ctx context.Context, ev cfn.Event, _ res1Input) (string, res1Output, bool, error) {
	return "", res1Output{}, false, nil
}

func (h hdl1) Update(ctx context.Context, ev cfn.Event, _, old res1Input) (string, res1Output, bool, error) {
	return "", res1Output{}, false, nil
}

func (h hdl1) Delete(ctx context.Context, ev cfn.Event, _ res1Input) (res1Output, bool, error) {
	return res1Output{}, false, nil
}

// hdl2.
type hdl2 struct{}

type (
	res2Input struct {
		Foo string `validate:"hostname"`
	}
	res2Output struct {
		Bar string
	}
)

func (h hdl2) Type() string { return "Custom::Hdl2" }
func (h hdl2) Create(ctx context.Context, ev cfn.Event, newi res2Input) (string, res2Output, bool, error) {
	return "", res2Output{Bar: newi.Foo + newi.Foo + "created"}, false, nil
}

func (h hdl2) Update(ctx context.Context, ev cfn.Event, newi, old res2Input) (string, res2Output, bool, error) {
	return "", res2Output{Bar: newi.Foo + old.Foo + "updated"}, false, nil
}

func (h hdl2) Delete(ctx context.Context, ev cfn.Event, newi res2Input) (res2Output, bool, error) {
	return res2Output{Bar: newi.Foo + newi.Foo + "deleted"}, false, nil
}

// hdl3.
type hdl3 struct{}

type (
	res3Input  struct{}
	res3Output struct{}
)

func (h hdl3) Type() string { return "Custom::Hdl3" }
func (h hdl3) Create(ctx context.Context, ev cfn.Event, _ res3Input) (string, res3Output, bool, error) {
	return "", res3Output{}, false, errors.New("some error")
}

func (h hdl3) Update(ctx context.Context, ev cfn.Event, _, old res3Input) (string, res3Output, bool, error) {
	return "", res3Output{}, false, nil
}

func (h hdl3) Delete(ctx context.Context, ev cfn.Event, _ res3Input) (res3Output, bool, error) {
	return res3Output{}, false, nil
}
