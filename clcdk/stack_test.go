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
		app.Node().SetContext(jsii.String("instance"), jsii.String("1"))

		conv = clcdk.NewConventions("ClFoo", "eu-west-1")
	})

	It("should create an instanced stack", func() {
		stack := clcdk.NewInstancedStack(app, conv, "1111111111")

		tmpl := assertions.Template_FromStack(stack, nil)
		data := *tmpl.ToJSON()

		Expect(data["Description"]).To(Equal("ClFoo (instance: 1)"))
	})
})
