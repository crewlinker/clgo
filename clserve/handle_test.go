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
	someJson := []byte(`[{"_id":"63dce6aef7179ac20c5836ca","index":0,"guid":"e1d5dcf8-074c-4716-92a1-da2eda781a43","tags":["aliquip","do","reprehenderit","sunt","amet","tempor","sit"],"friends":[{"id":0,"name":"Boyd Barnett"},{"id":1,"name":"Malone David"},{"id":2,"name":"Taylor Drake"}],"greeting":"Hello, undefined! You have 2 unread messages.","favoriteFruit":"apple"},{"_id":"63dce6ae6a88ed86f274529a","index":1,"guid":"46f5d4c7-502d-4cf2-a5e5-48966daf8d55","tags":["laborum","duis","incididunt","cupidatat","excepteur","nulla","laboris"],"friends":[{"id":0,"name":"Black Stevens"},{"id":1,"name":"Norris Herman"},{"id":2,"name":"Jocelyn Foster"}],"greeting":"Hello, undefined! You have 5 unread messages.","favoriteFruit":"apple"},{"_id":"63dce6aeb18a37f4439816ce","index":2,"guid":"455a3bea-c30d-4221-ab72-2b5a8111db33","tags":["duis","ut","non","non","cillum","mollit","enim"],"friends":[{"id":0,"name":"Coffey Thomas"},{"id":1,"name":"Augusta Vega"},{"id":2,"name":"Pauline Harmon"}],"greeting":"Hello, undefined! You have 9 unread messages.","favoriteFruit":"banana"},{"_id":"63dce6aeaf886da05db5f15b","index":3,"guid":"27d0a49f-aa59-4350-b75d-e4422d53aa50","tags":["dolore","qui","cupidatat","irure","deserunt","consectetur","reprehenderit"],"friends":[{"id":0,"name":"Loraine Sykes"},{"id":1,"name":"Leslie Armstrong"},{"id":2,"name":"Mason Davis"}],"greeting":"Hello, undefined! You have 1 unread messages.","favoriteFruit":"apple"},{"_id":"63dce6aed3a2a05d306a8423","index":4,"guid":"29fb8778-3254-4429-bbc0-cb8ca2cc2585","tags":["magna","minim","cupidatat","qui","dolore","id","labore"],"friends":[{"id":0,"name":"Aida Fowler"},{"id":1,"name":"Vinson Cohen"},{"id":2,"name":"Fitzgerald Cantu"}],"greeting":"Hello, undefined! You have 2 unread messages.","favoriteFruit":"apple"},{"_id":"63dce6ae203986f9221ea262","index":5,"guid":"1fae152c-d977-4858-879c-4e9b9ad3a291","tags":["nisi","minim","veniam","ullamco","enim","qui","reprehenderit"],"friends":[{"id":0,"name":"Reese Deleon"},{"id":1,"name":"Charity Sargent"},{"id":2,"name":"Gracie Rivera"}],"greeting":"Hello, undefined! You have 9 unread messages.","favoriteFruit":"banana"}]`)

	for _, c := range []struct {
		name string
		hf   func(http.ResponseWriter, *http.Request)
	}{
		{name: "one byte", hf: func(w http.ResponseWriter, r *http.Request) { w.Write([]byte{0x01}) }}, // one byte
		{name: "64KiB", hf: func(w http.ResponseWriter, r *http.Request) { w.Write(data64Kib) }},       // write 64KiB
		{name: "some json", hf: func(w http.ResponseWriter, r *http.Request) { w.Write(someJson) }},    // some json
	} {
		b.Run("buffered-"+c.name, func(b *testing.B) {
			h := clserve.Handle(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				c.hf(w, r)
				return nil
			})

			b.ResetTimer()
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				r, w := httptest.NewRequest("GET", "/", nil), httptest.NewRecorder()
				h.ServeHTTP(w, r)
			}
		})

		b.Run("original-"+c.name, func(b *testing.B) {
			h := http.HandlerFunc(c.hf)
			b.ResetTimer()
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				r, w := httptest.NewRequest("GET", "/", nil), httptest.NewRecorder()
				h.ServeHTTP(w, r)
			}
		})
	}
}

type MyContext struct {
	context.Context
	Foo string
}

type failingResponseWriter struct{ http.ResponseWriter }

func (failingResponseWriter) Write([]byte) (int, error) {
	return 0, errors.New("write fail")
}

var _ = Describe("handle implementations", func() {
	It("should return usual response on success", func(ctx context.Context) {
		w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)
		clserve.Handle(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			w.Header().Set("X-Foo", "bar")
			w.WriteHeader(201)
			fmt.Fprintf(w, "show this")
			return nil
		}).ServeHTTP(w, r)

		Expect(w).To(HaveHTTPStatus(201))
		Expect(w).To(HaveHTTPBody("show this"))
		Expect(w).To(HaveHTTPHeaderWithValue("X-Foo", "bar"))
	})

	Describe("error handling", func() {
		It("should return complete new error response on error", func(ctx context.Context) {
			w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)
			clserve.Handle(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				w.Header().Set("X-Foo", "bar")
				w.WriteHeader(201)
				fmt.Fprintf(w, "discard this")

				return errors.New("some error")
			}).ServeHTTP(w, r)

			Expect(w).To(HaveHTTPStatus(500))
			Expect(w).To(HaveHTTPBody("Internal Server Error\n"))
			Expect(w).ToNot(HaveHTTPHeaderWithValue("X-Foo", "bar"))
		})

		It("should allow custom error handler", func(ctx context.Context) {
			errh := func(c context.Context, w http.ResponseWriter, r *http.Request, e error) {
				w.Header().Set("X-Error", "my-error")
				w.WriteHeader(505)
				fmt.Fprintf(w, "my error")
			}

			w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)
			clserve.Handle(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				w.Header().Set("X-Foo", "bar")
				w.WriteHeader(201)
				fmt.Fprintf(w, "discard this")
				return errors.New("failed")
			}, clserve.WithErrorHandling(errh)).ServeHTTP(w, r)

			Expect(w).To(HaveHTTPStatus(505))
			Expect(w).To(HaveHTTPBody("my error"))
			Expect(w).To(HaveHTTPHeaderWithValue("X-Error", "my-error"))
		})
	})

	Describe("context building", func() {
		It("should allow custom context building", func(ctx context.Context) {
			ctxb := func(r *http.Request) (*MyContext, error) {
				return &MyContext{Foo: "bar"}, nil
			}

			w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)
			clserve.Handle(func(ctx *MyContext, w http.ResponseWriter, r *http.Request) error {
				fmt.Fprintf(w, ctx.Foo)
				return nil
			}, clserve.WithContextBuilder(ctxb)).ServeHTTP(w, r)

			Expect(w).To(HaveHTTPBody("bar"))
		})

		It("should handle error from ctx building", func(ctx context.Context) {
			w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)
			errh := func(c *MyContext, w http.ResponseWriter, r *http.Request, e error) {
				fmt.Fprintf(w, "%v", e)
			}
			ctxb := func(r *http.Request) (*MyContext, error) {
				return nil, fmt.Errorf("ctx fail")
			}

			clserve.Handle(func(ctx *MyContext, w http.ResponseWriter, r *http.Request) error {
				return nil
			}, clserve.WithContextBuilder(ctxb), clserve.WithErrorHandling(errh)).ServeHTTP(w, r)
			Expect(w).To(HaveHTTPBody("ctx fail"))
		})

		It("should panic without custom context builder", func(ctx context.Context) {
			Expect(func() {
				w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)
				clserve.Handle(func(ctx *MyContext, w http.ResponseWriter, r *http.Request) error {
					return nil
				}).ServeHTTP(w, r)
			}).Should(PanicWith(clserve.ErrContextBuilderRequired))
		})
	})

	It("error on writing to limit buffer", func(ctx context.Context) {
		w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)
		var errIsFull bool
		clserve.Handle(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			_, err := fmt.Fprintf(w, "xy")
			errIsFull = errors.Is(err, clserve.ErrBufferFull)
			return nil
		}, clserve.WithBufferLimit[context.Context](1)).ServeHTTP(w, r)
		Expect(errIsFull).To(BeTrue())
	})

	It("should handle reset errors", func(ctx context.Context) {
		w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)
		var errIsAlreadyFlushed bool
		errh := func(c context.Context, w http.ResponseWriter, r *http.Request, e error) {
			errIsAlreadyFlushed = errors.Is(e, clserve.ErrAlreadyFlushed)
		}

		clserve.Handle(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			rc := http.NewResponseController(w)
			Expect(rc.Flush()).To(Succeed())
			// return an error to trigger a reset on the buffered response, which will fail since
			// it's already flushed.
			return errors.New("foo")
		}, clserve.WithErrorHandling(errh)).ServeHTTP(w, r)
		Expect(errIsAlreadyFlushed).To(BeTrue())
	})

	It("should handle flushing error", func(ctx context.Context) {
		w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)
		fw := failingResponseWriter{w}
		var err error
		errh := func(c context.Context, w http.ResponseWriter, r *http.Request, e error) {
			err = e
		}

		clserve.Handle(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			fmt.Fprintf(w, "some data")
			return nil
		}, clserve.WithErrorHandling(errh)).ServeHTTP(fw, r)
		Expect(err).To(MatchError(`write fail`))
	})
})
