package clservev2_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/crewlinker/clgo/clservev2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("middleware", func() {
	It("should just return the handler without middleware", func() {
		hdlr1 := clservev2.HandlerFunc[context.Context](func(context.Context, clservev2.ResponseWriter, *http.Request) error {
			return nil
		})

		hdlr2 := clservev2.Use[context.Context](hdlr1)
		Expect(fmt.Sprint(hdlr1)).To(Equal(fmt.Sprint(hdlr2))) // compare addrs
	})

	It("should wrap in the correct order, and allow context to be modif", func() {
		var res string
		hdlr1 := clservev2.HandlerFunc[MyContext](func(MyContext, clservev2.ResponseWriter, *http.Request) error {
			res += "inner"

			return errors.New("inner error")
		})

		mw1 := func(n clservev2.Handler[MyContext]) clservev2.Handler[MyContext] {
			return clservev2.HandlerFunc[MyContext](func(c MyContext, w clservev2.ResponseWriter, r *http.Request) error {
				c.Foo = "some value"

				res += "1("
				err := n.ServeHTTP(c, w, r)
				res += ")1"

				return fmt.Errorf("1(%w)", err)
			})
		}

		mw2 := func(n clservev2.Handler[MyContext]) clservev2.Handler[MyContext] {
			return clservev2.HandlerFunc[MyContext](func(c MyContext, w clservev2.ResponseWriter, r *http.Request) error {
				Expect(c.Foo).ToNot(BeEmpty())

				res += "2("
				err := n.ServeHTTP(c, w, r)
				res += ")2"

				return fmt.Errorf("2(%w)", err)
			})
		}

		mw3 := func(n clservev2.Handler[MyContext]) clservev2.Handler[MyContext] {
			return clservev2.HandlerFunc[MyContext](func(c MyContext, w clservev2.ResponseWriter, r *http.Request) error {
				res += "3("
				err := n.ServeHTTP(c, w, r)
				res += ")3"

				return fmt.Errorf("3(%w)", err)
			})
		}

		var c MyContext
		err := clservev2.Use(hdlr1, mw1, mw2, mw3).ServeHTTP(c, nil, nil)
		Expect(res).To(Equal("1(2(3(inner)3)2)1"))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(`1(2(3(inner error)))`))
	})

	It("should panic, recover, reset the response and return a new error response", func(ctx context.Context) {
		hdlr1 := clservev2.Use(
			clservev2.HandlerFunc[context.Context](func(_ context.Context, w clservev2.ResponseWriter, r *http.Request) error {
				w.Header().Set("X-Foo", "bar")
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintf(w, "some body")

				panic("some panic")
			}),
			clservev2.Errorer[context.Context](),
			clservev2.Recoverer[context.Context](),
		)

		rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
		clservev2.Serve(hdlr1).ServeHTTP(rec, req)

		Expect(rec.Header()).To(Equal(http.Header{
			"Content-Type":           {"text/plain; charset=utf-8"},
			"X-Content-Type-Options": {"nosniff"},
		}))
		Expect(rec.Body.String()).To(Equal(`recovered: some panic` + "\n"))
	})
})
