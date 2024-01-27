package clcdk_test

import (
	"os"

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

	Describe("V1", func() {
		It("should create an instanced stack with instance context", func() {
			app.Node().SetContext(jsii.String("instance"), jsii.String("1"))
			app.Node().SetContext(jsii.String("environment"), jsii.String("dev"))

			stack := clcdk.NewInstancedStackV1(app, conv)
			tmpl := assertions.Template_FromStack(stack, nil)
			data := *tmpl.ToJSON()

			Expect(data["Description"]).To(Equal("ClFoo (env: dev, instance: 1)"))
		})

		It("should create a singleton stack with instance context", func() {
			app.Node().SetContext(jsii.String("environment"), jsii.String("dev"))

			stack := clcdk.NewSingletonStackV1(app, conv)
			tmpl := assertions.Template_FromStack(stack, nil)
			data := *tmpl.ToJSON()

			Expect(data["Description"]).To(Equal("ClFoo (env: dev, singleton)"))
			Expect(*awscdk.Stack_Of(stack).Account()).To(Equal("111111"))
		})

		// we don't want the code to panic without an instance or the bootstrap logic won't succeed. In
		// case of the bootstrap we never have an instance in the context.
		It("should not panic without instance context", func() {
			stack := clcdk.NewInstancedStackV1(app, conv)

			tmpl := assertions.Template_FromStack(stack, nil)
			data := *tmpl.ToJSON()

			Expect(data["Description"]).To(Equal("ClFoo (env: <none>, instance: 0)"))
			Expect(*awscdk.Stack_Of(stack).Account()).To(Equal("111111"))
		})
	})

	Describe("new instanced", Serial, func() {
		BeforeEach(func() {
			os.Setenv("CDK_DEFAULT_REGION", "eu-foo-1")
			os.Setenv("CDK_DEFAULT_ACCOUNT", "1111111")
		})

		It("should create an instanced stack with instance context", func() {
			app.Node().SetContext(jsii.String("qualifier"), jsii.String("ClFoo"))
			app.Node().SetContext(jsii.String("instance"), jsii.String("1"))
			app.Node().SetContext(jsii.String("environment"), jsii.String("dev"))

			stack := clcdk.NewInstancedStack(app)
			tmpl := assertions.Template_FromStack(stack, nil)
			data := *tmpl.ToJSON()

			Expect(data["Description"]).To(Equal("ClFoo (env: dev, instance: 1)"))
			Expect(*awscdk.Stack_Of(stack).Account()).To(Equal("1111111"))
			Expect(*awscdk.Stack_Of(stack).Region()).To(Equal("eu-foo-1"))
		})

		// we don't want the code to panic without an instance or the bootstrap logic won't succeed. In
		// case of the bootstrap we never have an instance in the context.
		It("should not panic without instance context", func() {
			app.Node().SetContext(jsii.String("qualifier"), jsii.String("ClFoo"))

			stack := clcdk.NewInstancedStack(app)

			tmpl := assertions.Template_FromStack(stack, nil)
			data := *tmpl.ToJSON()

			Expect(data["Description"]).To(Equal("ClFoo (env: , instance: 0)"))
			Expect(*awscdk.Stack_Of(stack).Account()).To(Equal("1111111"))
			Expect(*awscdk.Stack_Of(stack).Region()).To(Equal("eu-foo-1"))
		})
	})

	Describe("new regional singleton", Serial, func() {
		BeforeEach(func() {
			os.Setenv("CDK_DEFAULT_REGION", "eu-foo-1")
			os.Setenv("CDK_DEFAULT_ACCOUNT", "2222222")
		})

		It("should create an instanced stack with instance context", func() {
			app.Node().SetContext(jsii.String("qualifier"), jsii.String("ClFoo"))
			app.Node().SetContext(jsii.String("environment"), jsii.String("dev"))

			stack := clcdk.NewRegionalSingletonStack(app, "eu-bar-2", "EUB")
			tmpl := assertions.Template_FromStack(stack, nil)
			data := *tmpl.ToJSON()

			Expect(*stack.Node().Id()).To(Equal(`ClFooEUB`))
			Expect(data["Description"]).To(Equal("ClFoo (env: dev, singleton, eu-bar-2)"))
			Expect(*awscdk.Stack_Of(stack).Account()).To(Equal("2222222"))
			Expect(*awscdk.Stack_Of(stack).Region()).To(Equal("eu-bar-2"))
		})
	})

	Describe("new regional instanced", Serial, func() {
		BeforeEach(func() {
			os.Setenv("CDK_DEFAULT_REGION", "eu-foo-1")
			os.Setenv("CDK_DEFAULT_ACCOUNT", "2222222")
		})

		It("should create an instanced stack with instance context", func() {
			app.Node().SetContext(jsii.String("qualifier"), jsii.String("ClFoo"))
			app.Node().SetContext(jsii.String("instance"), jsii.String("2"))
			app.Node().SetContext(jsii.String("environment"), jsii.String("dev"))

			stack := clcdk.NewRegionalInstancedStack(app, "eu-bar-2", "EUB")
			tmpl := assertions.Template_FromStack(stack, nil)
			data := *tmpl.ToJSON()

			Expect(*stack.Node().Id()).To(Equal(`ClFooEUB2`))
			Expect(data["Description"]).To(Equal("ClFoo (env: dev, instance: 2, eu-bar-2)"))
			Expect(*awscdk.Stack_Of(stack).Account()).To(Equal("2222222"))
			Expect(*awscdk.Stack_Of(stack).Region()).To(Equal("eu-bar-2"))
		})
	})
})
