package clcdk_test

import (
	"encoding/json"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/jsii-runtime-go"
	"github.com/crewlinker/clgo/clcdk"
	"github.com/samber/lo"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("refs", func() {
	var app awscdk.App

	BeforeEach(func() {
		app = awscdk.NewApp(nil)
	})

	It("export import", func() {
		By("exporting it from a stack")
		stack1 := awscdk.NewStack(app, jsii.String("Stack1"), nil)
		bucket1 := awss3.NewBucket(stack1, jsii.String("Bucket1"), nil)

		ref1 := clcdk.ExportValue(stack1, bucket1.BucketName())

		By("importing it in another stack")
		stack2 := awscdk.NewStack(app, jsii.String("Stack2"), nil)
		stack2.ExportValue(ref1.ImportValue(), &awscdk.ExportValueOptions{Name: jsii.String("ReExport1")})

		By("asserting templates stack1's output")
		tmpl1 := assertions.Template_FromStack(stack1, nil)
		map1 := *tmpl1.ToJSON()
		Expect(map1["Outputs"]).To(HaveLen(1))
		json1 := lo.Must(json.Marshal(map1["Outputs"]))
		Expect(json1).To(ContainSubstring(`{"Export":{"Name":"Stack1:ExportsOutputRefBucket`))

		tmpl2 := assertions.Template_FromStack(stack2, nil)
		map2 := *tmpl2.ToJSON()
		json2 := lo.Must(json.Marshal(map2))
		Expect(json2).To(ContainSubstring(`"Fn::ImportValue":"Stack1:ExportsOutputRef`))

		Expect(strings.Count(string(json2), "Fn::ImportValue")).To(Equal(1))
	})
})
