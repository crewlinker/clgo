package clserve_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/crewlinker/clgo/clserve"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func BenchmarkHandler(b *testing.B) {
	data64Kib := make([]byte, 1024*64)
	someJSON := []byte(`[{"_id":"63dce6aef7179ac20c5836ca","index":0,"guid":"e1d5dcf8-074c-4716-92a1-da2eda781a43","tags":["aliquip","do","reprehenderit","sunt","amet","tempor","sit"],"friends":[{"id":0,"name":"Boyd Barnett"},{"id":1,"name":"Malone David"},{"id":2,"name":"Taylor Drake"}],"greeting":"Hello, undefined! You have 2 unread messages.","favoriteFruit":"apple"},{"_id":"63dce6ae6a88ed86f274529a","index":1,"guid":"46f5d4c7-502d-4cf2-a5e5-48966daf8d55","tags":["laborum","duis","incididunt","cupidatat","excepteur","nulla","laboris"],"friends":[{"id":0,"name":"Black Stevens"},{"id":1,"name":"Norris Herman"},{"id":2,"name":"Jocelyn Foster"}],"greeting":"Hello, undefined! You have 5 unread messages.","favoriteFruit":"apple"},{"_id":"63dce6aeb18a37f4439816ce","index":2,"guid":"455a3bea-c30d-4221-ab72-2b5a8111db33","tags":["duis","ut","non","non","cillum","mollit","enim"],"friends":[{"id":0,"name":"Coffey Thomas"},{"id":1,"name":"Augusta Vega"},{"id":2,"name":"Pauline Harmon"}],"greeting":"Hello, undefined! You have 9 unread messages.","favoriteFruit":"banana"},{"_id":"63dce6aeaf886da05db5f15b","index":3,"guid":"27d0a49f-aa59-4350-b75d-e4422d53aa50","tags":["dolore","qui","cupidatat","irure","deserunt","consectetur","reprehenderit"],"friends":[{"id":0,"name":"Loraine Sykes"},{"id":1,"name":"Leslie Armstrong"},{"id":2,"name":"Mason Davis"}],"greeting":"Hello, undefined! You have 1 unread messages.","favoriteFruit":"apple"},{"_id":"63dce6aed3a2a05d306a8423","index":4,"guid":"29fb8778-3254-4429-bbc0-cb8ca2cc2585","tags":["magna","minim","cupidatat","qui","dolore","id","labore"],"friends":[{"id":0,"name":"Aida Fowler"},{"id":1,"name":"Vinson Cohen"},{"id":2,"name":"Fitzgerald Cantu"}],"greeting":"Hello, undefined! You have 2 unread messages.","favoriteFruit":"apple"},{"_id":"63dce6ae203986f9221ea262","index":5,"guid":"1fae152c-d977-4858-879c-4e9b9ad3a291","tags":["nisi","minim","veniam","ullamco","enim","qui","reprehenderit"],"friends":[{"id":0,"name":"Reese Deleon"},{"id":1,"name":"Charity Sargent"},{"id":2,"name":"Gracie Rivera"}],"greeting":"Hello, undefined! You have 9 unread messages.","favoriteFruit":"banana"}]`)

	for _, tblc := range []struct {
		name string
		hf   func(http.ResponseWriter, *http.Request)
	}{
		{name: "one byte", hf: func(w http.ResponseWriter, r *http.Request) {
			Expect(w.Write([]byte{0x01})).NotTo(BeZero())
		}}, // one byte
		{name: "64KiB", hf: func(w http.ResponseWriter, r *http.Request) {
			Expect(w.Write(data64Kib)).NotTo(BeZero())
		}}, // write 64KiB
		{name: "some json", hf: func(w http.ResponseWriter, r *http.Request) {
			Expect(w.Write(someJSON)).NotTo(BeZero())
		}}, // some json
	} {
		b.Run("buffered-"+tblc.name, func(b *testing.B) {
			hdlr := clserve.Handle(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				tblc.hf(w, r)

				return nil
			})

			b.ResetTimer()
			b.ReportAllocs()

			for n := 0; n < b.N; n++ {
				r, w := httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder()
				hdlr.ServeHTTP(w, r)
			}
		})

		b.Run("original-"+tblc.name, func(b *testing.B) {
			h := http.HandlerFunc(tblc.hf)

			b.ResetTimer()
			b.ReportAllocs()

			for n := 0; n < b.N; n++ {
				r, w := httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder()
				h.ServeHTTP(w, r)
			}
		})
	}
}

type MyContext struct {
	context.Context //nolint:containedctx
	Foo             string
}

type failingResponseWriter struct{ http.ResponseWriter }

var (
	errWriteFail = errors.New("write fail")
	errSome      = errors.New("some error")
	errFailed    = errors.New("failed")
	errFoo       = errors.New("foo")
	errCtxFail   = errors.New("ctx fail")
)

func (failingResponseWriter) Write([]byte) (int, error) {
	return 0, errWriteFail
}

var _ = Describe("handle implementations", func() {
	It("should return usual response on success", func(ctx context.Context) {
		resp, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
		clserve.Handle(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			w.Header().Set("X-Foo", "bar")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, "show this")

			return nil
		}).ServeHTTP(resp, req)

		Expect(resp).To(HaveHTTPStatus(201))
		Expect(resp).To(HaveHTTPBody("show this"))
		Expect(resp).To(HaveHTTPHeaderWithValue("X-Foo", "bar"))
	})

	Describe("error handling", func() {
		It("should return complete new error response on error", func(ctx context.Context) {
			resp, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			clserve.Handle(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				w.Header().Set("X-Foo", "bar")
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintf(w, "discard this")

				return errSome
			}).ServeHTTP(resp, req)

			Expect(resp).To(HaveHTTPStatus(500))
			Expect(resp).To(HaveHTTPBody("Internal Server Error\n"))
			Expect(resp).ToNot(HaveHTTPHeaderWithValue("X-Foo", "bar"))
		})

		It("should allow custom error handler", func(ctx context.Context) {
			errh := func(c context.Context, w http.ResponseWriter, r *http.Request, e error) {
				w.Header().Set("X-Error", "my-error")
				w.WriteHeader(http.StatusHTTPVersionNotSupported)
				fmt.Fprintf(w, "my error")
			}

			resp, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			clserve.Handle(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				w.Header().Set("X-Foo", "bar")
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintf(w, "discard this")

				return errFailed
			}, clserve.WithContextErrorHandling(errh)).ServeHTTP(resp, req)

			Expect(resp).To(HaveHTTPStatus(505))
			Expect(resp).To(HaveHTTPBody("my error"))
			Expect(resp).To(HaveHTTPHeaderWithValue("X-Error", "my-error"))
		})

		It("hould handle panics through error handler by default", func(ctx context.Context) {
			errh := func(c context.Context, w http.ResponseWriter, r *http.Request, e error) {
				fmt.Fprintf(w, "%v", e)
			}

			resp, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			clserve.Handle(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				panic("my panic")
			}, clserve.WithContextErrorHandling(errh)).ServeHTTP(resp, req)

			Expect(resp).To(HaveHTTPStatus(200))
			Expect(resp).To(HaveHTTPBody("my panic"))
		})

		It("should allow custom panic handling", func(ctx context.Context) {
			panich := func(
				c context.Context, w http.ResponseWriter, r *http.Request, v any, errh clserve.ErrorHandlerFunc[context.Context],
			) {
				w.WriteHeader(http.StatusBadGateway)
				fmt.Fprintf(w, "panic: %v", v)
			}

			resp, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			clserve.Handle(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				panic("my panic")
			}, clserve.WithPanicHandler(panich)).ServeHTTP(resp, req)

			Expect(resp).To(HaveHTTPStatus(502))
			Expect(resp).To(HaveHTTPBody("panic: my panic"))
		})

		It("should allow disabling of panic handling", func(ctx context.Context) {
			Expect(func() {
				w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
				clserve.Handle(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
					panic("my panic")
				}, clserve.WithPanicHandler[context.Context](nil)).ServeHTTP(w, r)
			}).To(Panic())
		})
	})

	Describe("context building", func() {
		It("should allow custom context building", func(ctx context.Context) {
			ctxb := func(r *http.Request) (*MyContext, error) {
				return &MyContext{Foo: "bar"}, nil
			}

			resp, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			clserve.Handle(func(ctx *MyContext, w http.ResponseWriter, r *http.Request) error {
				fmt.Fprintf(w, ctx.Foo)

				return nil
			}, clserve.WithContextBuilder(ctxb)).ServeHTTP(resp, req)

			Expect(resp).To(HaveHTTPBody("bar"))
		})

		It("should handle error from ctx building", func(ctx context.Context) {
			resp, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
			errh := func(w http.ResponseWriter, r *http.Request, e error) {
				fmt.Fprintf(w, "%v", e)
			}
			ctxb := func(r *http.Request) (*MyContext, error) {
				return nil, errCtxFail
			}

			clserve.Handle(func(ctx *MyContext, w http.ResponseWriter, r *http.Request) error {
				return nil
			}, clserve.WithContextBuilder(ctxb), clserve.WithErrorHandling[*MyContext](errh)).ServeHTTP(resp, req)
			Expect(resp).To(HaveHTTPBody("ctx fail"))
		})

		It("should panic without custom context builder", func(ctx context.Context) {
			Expect(func() {
				w, r := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
				clserve.Handle(func(ctx *MyContext, w http.ResponseWriter, r *http.Request) error {
					return nil
				}).ServeHTTP(w, r)
			}).Should(PanicWith(clserve.ErrContextBuilderRequired))
		})
	})

	It("error on writing to limit buffer", func(ctx context.Context) {
		resp, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
		var errIsFull bool
		clserve.Handle(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			_, err := fmt.Fprintf(w, "xy")
			errIsFull = errors.Is(err, clserve.ErrBufferFull)

			return nil
		}, clserve.WithBufferLimit[context.Context](1)).ServeHTTP(resp, req)
		Expect(errIsFull).To(BeTrue())
	})

	It("should handle reset errors", func(ctx context.Context) {
		resp, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
		var errIsAlreadyFlushed bool
		errh := func(c context.Context, w http.ResponseWriter, r *http.Request, e error) {
			errIsAlreadyFlushed = errors.Is(e, clserve.ErrAlreadyFlushed)
		}

		clserve.Handle(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			rc := http.NewResponseController(w)
			Expect(rc.Flush()).To(Succeed())
			// return an error to trigger a reset on the buffered response, which will fail since
			// it's already flushed.
			return errFoo
		}, clserve.WithContextErrorHandling(errh)).ServeHTTP(resp, req)
		Expect(errIsAlreadyFlushed).To(BeTrue())
	})

	It("should handle flushing error", func(ctx context.Context) {
		resp, req := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil)
		fresp := failingResponseWriter{resp}
		var err error
		errh := func(c context.Context, w http.ResponseWriter, r *http.Request, e error) {
			err = e
		}

		clserve.Handle(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			fmt.Fprintf(w, "some data")

			return nil
		}, clserve.WithContextErrorHandling(errh)).ServeHTTP(fresp, req)

		Expect(err).To(MatchError(MatchRegexp(`write fail`)))
	})
})
