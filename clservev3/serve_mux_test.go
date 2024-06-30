package clservev3_test

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"

	"github.com/crewlinker/clgo/clservev3"
	"github.com/crewlinker/clgo/clservev3/internal/example"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("serve mux", func() {
	var mux *clservev3.ServeMux[TestValues]
	var testStdMiddleware clservev3.StdMiddleware

	BeforeEach(func() {
		testStdMiddleware = func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), "ctxv1", "bar") //nolint:staticcheck

				next.ServeHTTP(w, r.WithContext(ctx))
			})
		}

		mux = clservev3.NewServeMux[TestValues]()
		mux.Use(testStdMiddleware)
		mux.BUse(example.Middleware[TestValues](slog.Default()))
		mux.BHandleFunc("GET /blog/{slug}", func(ctx *clservev3.Context[TestValues], w clservev3.ResponseWriter, r *http.Request) error {
			Expect(ctx.V.Logger).ToNot(BeNil())

			_, err := fmt.Fprintf(w, "%s: hello, %s (%v)", r.PathValue("slug"), r.RemoteAddr, r.Context().Value("ctxv1"))

			return err
		}, "blog_post")

		mux.HandleFunc("GET /blog/comment/{id}", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "comment %s: hello std, %s (%v)", r.PathValue("id"), r.RemoteAddr, r.Context().Value("ctxv1"))
		}, "blog_comment")
	})

	It("should reverse buffered", func() {
		reversed, err := mux.Reverse("blog_post", "slug2")
		Expect(err).ToNot(HaveOccurred())
		Expect(reversed).To(Equal(`/blog/slug2`))
	})

	It("should reverse standard", func() {
		reversed, err := mux.Reverse("blog_comment", "id1")
		Expect(err).ToNot(HaveOccurred())
		Expect(reversed).To(Equal(`/blog/comment/id1`))
	})

	It("should serve the buffered endpoint", func() {
		rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/blog/some-post", nil)
		mux.ServeHTTP(rec, req)

		Expect(rec.Result().StatusCode).To(Equal(http.StatusOK))
		Expect(rec.Body.String()).To(Equal(`some-post: hello, 192.0.2.1:1234 (bar)`))
	})

	It("should serve the standard endpoint", func() {
		rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/blog/comment/5", nil)
		mux.ServeHTTP(rec, req)

		Expect(rec.Result().StatusCode).To(Equal(http.StatusOK))
		Expect(rec.Body.String()).To(Equal(`comment 5: hello std, 192.0.2.1:1234 (bar)`))
	})
})
