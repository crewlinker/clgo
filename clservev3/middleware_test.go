package clservev3_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/crewlinker/clgo/clservev3"
	"github.com/crewlinker/clgo/clservev3/internal/example"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type TestValues struct {
	Foo    string
	Logger *slog.Logger
}

func (v TestValues) WithLogger(logs *slog.Logger) TestValues {
	v.Logger = logs

	return v
}

var _ = Describe("middleware", func() {
	It("should just return the handler without middleware", func() {
		hdlr1 := clservev3.HandlerFunc[struct{}](func(*clservev3.Context[struct{}], clservev3.ResponseWriter, *http.Request) error {
			return nil
		})

		hdlr2 := clservev3.Use(hdlr1)
		Expect(fmt.Sprint(hdlr1)).To(Equal(fmt.Sprint(hdlr2))) // compare addrs
	})

	It("should wrap in the correct order, and allow context to be modif", func() {
		var res string
		hdlr1 := clservev3.HandlerFunc[TestValues](func(c *clservev3.Context[TestValues], _ clservev3.ResponseWriter, r *http.Request) error {
			res += fmt.Sprintf("inner %s %v", c.V.Foo, c.Value("foo"))

			By("making sure the request's context and c's carried data are equal")
			Expect(r.Context().Value("foo")).To(Equal(c.Value("foo")))

			By("making sure deadline is consistent between the two contexts")
			dl1, ok1 := c.Deadline()
			dl2, ok2 := r.Context().Deadline()
			Expect(dl1).To(Equal(dl2))
			Expect(ok1).To(Equal(ok2))

			Expect(c.V.Logger).ToNot(BeNil())

			return errors.New("inner error")
		})

		mw1 := func(n clservev3.Handler[TestValues]) clservev3.Handler[TestValues] {
			return clservev3.HandlerFunc[TestValues](func(c *clservev3.Context[TestValues], w clservev3.ResponseWriter, r *http.Request) error {
				res += "1("
				err := n.ServeBHTTP(c, w, r)
				res += ")1"

				return fmt.Errorf("1(%w)", err)
			})
		}

		mw2 := func(n clservev3.Handler[TestValues]) clservev3.Handler[TestValues] {
			return clservev3.HandlerFunc[TestValues](func(c *clservev3.Context[TestValues], w clservev3.ResponseWriter, r *http.Request) error {
				res += "2("
				err := n.ServeBHTTP(c, w, r)
				res += ")2"

				return fmt.Errorf("2(%w)", err)
			})
		}

		mw3 := func(n clservev3.Handler[TestValues]) clservev3.Handler[TestValues] {
			return clservev3.HandlerFunc[TestValues](func(c *clservev3.Context[TestValues], w clservev3.ResponseWriter, r *http.Request) error {
				c.V.Foo = "some value"

				c, r = c.WithValue(r, "foo", "bar")

				res += "3("
				err := n.ServeBHTTP(c, w, r)
				res += ")3"

				return fmt.Errorf("3(%w)", err)
			})
		}

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()

		rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
		req = req.WithContext(ctx)

		slog := slog.Default()

		bctx := clservev3.NewContext[TestValues](ctx)
		err := clservev3.Use(hdlr1, example.Middleware[TestValues](slog), mw3, mw2, mw1).ServeBHTTP(bctx, clservev3.NewBufferResponse(rec, -1), req)
		Expect(res).To(Equal("3(2(1(inner some value bar)1)2)3"))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(`3(2(1(inner error)))`))
	})

	It("should panic, recover, reset the response and return a new error response", func(ctx context.Context) {
		hdlr1 := clservev3.Use(
			clservev3.HandlerFunc[struct{}](func(_ *clservev3.Context[struct{}], w clservev3.ResponseWriter, r *http.Request) error {
				w.Header().Set("X-Foo", "bar")
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintf(w, "some body") // this will be reset

				panic("some panic")
			}),
			Errorer[struct{}](),
			Recoverer[struct{}](),
		)

		rec, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
		clservev3.Serve(hdlr1).ServeHTTP(rec, req)

		Expect(rec.Header()).To(Equal(http.Header{
			"Content-Type":           {"text/plain; charset=utf-8"},
			"X-Content-Type-Options": {"nosniff"},
		}))
		Expect(rec.Body.String()).To(Equal(`recovered: some panic` + "\n"))
	})
})

// Errorer middleware will reset the buffered response, and return a server error.
func Errorer[C any]() clservev3.Middleware[C] {
	return func(next clservev3.Handler[C]) clservev3.Handler[C] {
		return clservev3.HandlerFunc[C](func(c *clservev3.Context[C], w clservev3.ResponseWriter, r *http.Request) error {
			err := next.ServeBHTTP(c, w, r)
			if err != nil {
				w.Reset()
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			return nil
		})
	}
}

// Recover middleware. It will recover any panics and turn it into an error.
func Recoverer[C any]() clservev3.Middleware[C] {
	return func(next clservev3.Handler[C]) clservev3.Handler[C] {
		return clservev3.HandlerFunc[C](func(c *clservev3.Context[C], w clservev3.ResponseWriter, r *http.Request) (err error) {
			defer func() {
				if e := recover(); e != nil {
					err = fmt.Errorf("recovered: %v", e)
				}
			}()

			return next.ServeBHTTP(c, w, r)
		})
	}
}
