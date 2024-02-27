package clserve_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/crewlinker/clgo/clserve"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("errors", func() {
	It("should render coded error messages", func() {
		Expect(clserve.Errorf(404, "some message").Error()).To(Equal(`404: some message`))
	})

	Describe("show real errors", func() {
		var opts []clserve.Option[context.Context]
		BeforeEach(func() {
			opts = append(opts, clserve.WithErrorHandling[context.Context](
				clserve.StandarErrorHandler(true, http.Error),
			))
		})

		It("should do standard errors by default", func() {
			resp, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			clserve.Handle(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				return errors.New("some error")
			}, opts...).ServeHTTP(resp, req)

			Expect(resp).To(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(resp.Body.String()).To(Equal("some error\n"))
		})

		It("should do formatted errors", func() {
			resp, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			clserve.Handle(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				return clserve.Errorf(400, "some bad param: %d", 100)
			}, opts...).ServeHTTP(resp, req)

			Expect(resp).To(HaveHTTPStatus(http.StatusBadRequest))
			Expect(resp.Body.String()).To(Equal("some bad param: 100\n"))
		})
	})

	Describe("do not show real errors", func() {
		var opts []clserve.Option[context.Context]
		BeforeEach(func() {
			opts = append(opts, clserve.WithErrorHandling[context.Context](
				clserve.StandarErrorHandler(false, http.Error),
			))
		})

		It("should do standard errors by default", func() {
			resp, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			clserve.Handle(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				return errors.New("some error")
			}, opts...).ServeHTTP(resp, req)

			Expect(resp).To(HaveHTTPStatus(http.StatusInternalServerError))
			Expect(resp.Body.String()).To(Equal("Internal Server Error\n"))
		})

		It("should do formatted errors", func() {
			resp, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			clserve.Handle(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				return clserve.Errorf(400, "some bad param: %d", 100)
			}, opts...).ServeHTTP(resp, req)

			Expect(resp).To(HaveHTTPStatus(http.StatusBadRequest))
			Expect(resp.Body.String()).To(Equal("Bad Request\n"))
		})
	})
})
