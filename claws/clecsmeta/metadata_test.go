package clecsmeta_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/crewlinker/clgo/claws/clecsmeta"
	"github.com/crewlinker/clgo/clzap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"

	_ "embed"
)

//go:embed testdata/metadatav4_response_task.json
var exampleResponse1 []byte

func TestClecsmeta(t *testing.T) {
	t.Parallel()
	RegisterFailHandler(Fail)
	RunSpecs(t, "claws/clecsmeta")
}

var _ = Describe("metadata provide", func() {
	var meta *clecsmeta.Metadata

	BeforeEach(func(ctx context.Context) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, `{}`)
		}))

		os.Setenv("ECS_CONTAINER_METADATA_URI_V4", srv.URL)

		app := fx.New(fx.Populate(&meta), clecsmeta.Provide(), fx.Supply(http.DefaultClient), clzap.TestProvide())
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should di", func() {
		Expect(meta).ToNot(BeNil())
	})
})

var _ = Describe("metadata test provide", func() {
	var meta *clecsmeta.Metadata

	BeforeEach(func(ctx context.Context) {
		app := fx.New(fx.Populate(&meta), clecsmeta.TestProvide(), fx.Supply(http.DefaultClient), clzap.TestProvide())
		Expect(app.Start(ctx)).To(Succeed())
		DeferCleanup(app.Stop)
	})

	It("should setup", func() {
		Expect(meta).ToNot(BeNil())
		Expect(meta.TaskV4().Cluster).To(Equal(`default`))
	})

	It("should unmarshal", func() {
		var v clecsmeta.TaskMetadataV4
		Expect(json.Unmarshal(exampleResponse1, &v)).To(Succeed())

		Expect(v.Containers[1].LogOptions.AwsLogsGroup).To(Equal("/ecs/metadata"))
		Expect(v.Containers[1].LogOptions.AwsLogsStream).To(Equal("ecs/curl/158d1c8083dd49d6b527399fd6414f5c"))
		Expect(v.Containers[1].LogOptions.AwsRegion).To(Equal("us-west-2"))
		Expect(v.Containers[1].LogOptions.AwsLogsCreateGroup).To(Equal("true"))
		Expect(v.Containers[1].Labels.EcsTaskArn).To(Equal("arn:aws:ecs:us-west-2:111122223333:task/default/158d1c8083dd49d6b527399fd6414f5c"))
		Expect(v.Containers[1].Labels.Get("org.opencontainers.image.revision")).To(Equal("0xdeadbeaf"))
		Expect(v.Containers[1].Labels.Get("internal.not-exist-label")).To(Equal(""))
	})
})
