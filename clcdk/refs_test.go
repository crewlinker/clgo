package clcdk_test

import (
	"encoding/json"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
	"github.com/crewlinker/clgo/clcdk"
	"github.com/samber/lo"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("refs", func() {
	var app awscdk.App
	var conv clcdk.Conventions

	BeforeEach(func() {
		app = awscdk.NewApp(nil)
		conv = clcdk.NewConventions("ClFoo", "eu-west-1", "111111")
	})

	It("should create an instanced stack with instance context", func() {
		stack1 := awscdk.NewStack(app, jsii.String("Stack1"), nil)
		stack2 := awscdk.NewStack(app, jsii.String("Stack2"), nil)

		scope := clcdk.ScopeName("Scope1").ChildScope(stack1)
		scope = clcdk.ScopeName("Scope2").ChildScope(scope)

		ref1 := clcdk.Export(scope, "MyExport", conv, stack1.StackName())
		ref1.Import(stack2)

		tmpl1 := assertions.Template_FromStack(stack1, nil)
		map1 := *tmpl1.ToJSON()
		Expect(map1["Outputs"]).To(HaveLen(1))

		json1 := lo.Must(json.Marshal(map1["Outputs"]))
		Expect(json1).To(ContainSubstring(`"Description":"Stack1/Scope1/Scope2/MyExport"`))
		Expect(json1).To(ContainSubstring(`"Export":{"Name":"ClFooStack1Scope1Scope2MyExport"}`))
		Expect(json1).To(ContainSubstring(`"Value":"Stack1"`))

		Expect(*stack2.Dependencies()).To(HaveLen(1))
		Expect(*(*stack2.Dependencies())[0].StackName()).To(Equal("Stack1"))
	})
})
