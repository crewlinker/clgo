package clcdk_test

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/jsii-runtime-go"
	"github.com/crewlinker/clgo/clcdk"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("stack", func() {
	var app awscdk.App
	var conv clcdk.Conventions

	BeforeEach(func() {
		app = awscdk.NewApp(nil)

		conv = clcdk.NewConventions("ClFoo", "eu-west-1", "111111")
	})

	It("should create an instanced stack with instance context", func() {
		app.Node().SetContext(jsii.String("instance"), jsii.String("1"))
		app.Node().SetContext(jsii.String("environment"), jsii.String("dev"))

		stack := clcdk.NewInstancedStack(app, conv)
		tmpl := assertions.Template_FromStack(stack, nil)
		data := *tmpl.ToJSON()

		Expect(data["Description"]).To(Equal("ClFoo (env: dev, instance: 1)"))
	})

	It("should create a singleton stack with instance context", func() {
		app.Node().SetContext(jsii.String("environment"), jsii.String("dev"))

		stack := clcdk.NewSingletonStack(app, conv)
		tmpl := assertions.Template_FromStack(stack, nil)
		data := *tmpl.ToJSON()

		Expect(data["Description"]).To(Equal("ClFoo (env: dev, singleton)"))
		Expect(*awscdk.Stack_Of(stack).Account()).To(Equal("111111"))
	})

	// we don't want the code to panic without an instance or the bootstrap logic won't succeed. In
	// case of the bootstrap we never have an instance in the context.
	It("should not panic without instance context", func() {
		stack := clcdk.NewInstancedStack(app, conv)

		tmpl := assertions.Template_FromStack(stack, nil)
		data := *tmpl.ToJSON()

		Expect(data["Description"]).To(Equal("ClFoo (env: <none>, instance: 0)"))
		Expect(*awscdk.Stack_Of(stack).Account()).To(Equal("111111"))
	})
})
