package clserve_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/crewlinker/clgo/clserve"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func BenchmarkResponseBuffer(b *testing.B) {
	var w http.ResponseWriter
	for _, dat := range [][]byte{
		make([]byte, 1024),    // 1KiB
		make([]byte, 1024*64), // 64KiB
	} {
		b.Run("buffered-"+strconv.Itoa(len(dat)), func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				w = httptest.NewRecorder()
				w = clserve.NewBufferResponse(w, -1)
				w.Write(dat)
				w.(*clserve.ResponseBuffer).ImplicitFlush()
				w.(*clserve.ResponseBuffer).Free()
			}
		})
		b.Run("original-"+strconv.Itoa(len(dat)), func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				w = httptest.NewRecorder()
				w.Write(dat)
			}
		})
	}
}

var _ = Describe("handle implementations", func() {
	DescribeTable("equal", func(
		handler func(http.ResponseWriter, *http.Request),
		exp func(r1, r2 *http.Response, b1, b2 *bytes.Buffer),
	) {
		log1 := log.New(io.Discard, "", 0)
		log2 := log.New(io.Discard, "", 0)

		ln1, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).ToNot(HaveOccurred())
		ln2, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).ToNot(HaveOccurred())
		srv1 := &httptest.Server{
			Listener: ln1,
			Config:   &http.Server{Handler: http.HandlerFunc(handler), ErrorLog: log1},
		}
		srv1.Start()
		srv2 := &httptest.Server{
			Listener: ln2,
			Config: &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer GinkgoRecover()
				wr := clserve.NewBufferResponse(w, 10)
				handler(wr, r)
				Expect(wr.ImplicitFlush()).To(Succeed()) // flush any remaining data
			}), ErrorLog: log2},
		}
		srv2.Start()

		req1, _ := http.NewRequest("GET", srv1.URL, nil)
		req2, _ := http.NewRequest("GET", srv2.URL, nil)
		resp1, err1 := http.DefaultClient.Do(req1)
		resp2, err2 := http.DefaultClient.Do(req2)
		Expect(err1).ToNot(HaveOccurred())
		Expect(err2).ToNot(HaveOccurred())

		defer resp1.Body.Close()
		defer resp2.Body.Close()

		Expect(resp1.StatusCode).To(Equal(resp2.StatusCode))
		Expect(resp1.Header.Get("Rab")).To(Equal(resp2.Header.Get("Rab")))
		b1 := bytes.NewBuffer(nil)
		io.Copy(b1, resp1.Body)
		b2 := bytes.NewBuffer(nil)
		io.Copy(b2, resp2.Body)

		exp(resp1, resp2, b1, b2)
	},
		Entry("implicit 200",
			func(w http.ResponseWriter, r *http.Request) {},
			func(r1, r2 *http.Response, b1, b2 *bytes.Buffer) {
				Expect(r1.StatusCode).To(Equal(200))
			},
		),
		Entry("implicit header writing",
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("foo", "bar")
				fmt.Fprintf(w, "foo")
			},
			func(r1, r2 *http.Response, b1, b2 *bytes.Buffer) {
				Expect(r1.StatusCode).To(Equal(200))
				Expect(b1.String()).To(Equal("foo"))
			},
		),
		Entry("explicit 201",
			func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(201) },
			func(r1, r2 *http.Response, b1, b2 *bytes.Buffer) {
				Expect(r1.StatusCode).To(Equal(201))
			},
		),
		Entry("set header after .Write",
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Rab", "dar")
				fmt.Fprintf(w, "bar") // this should cause header to be flushed
				w.Header().Set("Dar", "tab")
			},
			func(r1, r2 *http.Response, b1, b2 *bytes.Buffer) {
				Expect(r1.StatusCode).To(Equal(200))
				Expect(r1.Header.Get("Rab")).To(Equal("dar"))
			},
		),
		Entry("set header after .WriteHeader",
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Rab", "dar")
				w.WriteHeader(202)
				w.Header().Set("Dar", "tab")
			},
			func(r1, r2 *http.Response, b1, b2 *bytes.Buffer) {
				Expect(r1.StatusCode).To(Equal(202))
				Expect(r1.Header.Get("Rab")).To(Equal("dar"))
			},
		),
		Entry("headers without explicit .Write or WriteHeader",
			func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Rab", "dar")
				w.Header().Set("Dar", "tab")
			},
			func(r1, r2 *http.Response, b1, b2 *bytes.Buffer) {
				Expect(r1.StatusCode).To(Equal(200))
				Expect(r1.Header.Get("Rab")).To(Equal("dar"))
				Expect(r1.Header.Get("Dar")).To(Equal("tab"))
			},
		),
		Entry("headers and body after explicit Flush",
			func(w http.ResponseWriter, r *http.Request) {
				rc := http.NewResponseController(w)
				w.Header().Set("Rab", "dar")
				fmt.Fprintf(w, "aaa")
				Expect(rc.Flush()).To(Succeed())
				w.Header().Set("Dar", "tab")
			},
			func(r1, r2 *http.Response, b1, b2 *bytes.Buffer) {
				Expect(r1.StatusCode).To(Equal(200))
				Expect(r1.Header.Get("Rab")).To(Equal("dar"))
				Expect(b1.String()).To(Equal("aaa"))
			},
		),
		Entry("set status after write",
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(201)
				fmt.Fprintf(w, "bar")
				w.WriteHeader(202)
			},
			func(r1, r2 *http.Response, b1, b2 *bytes.Buffer) {
				Expect(r1.StatusCode).To(Equal(201))
				Expect(b1.String()).To(Equal("bar"))
			},
		),
	)
})

var _ = Describe("buffered writes", func() {
	var wr io.Writer
	var fwr interface {
		io.Writer                    // should always be a regular writer
		FlushError() error           // implemented for: https://pkg.go.dev/net/http@master#NewResponseController
		Unwrap() http.ResponseWriter // implemented for: https://pkg.go.dev/net/http@master#NewResponseController
	}

	var resp interface {
		http.ResponseWriter
		Reset() error
		FlushError() error
	}

	Describe("limiting", func() {
		It("should limit writes exactly", func() {
			rec := httptest.NewRecorder()
			wr = clserve.NewBufferResponse(rec, 1)
			Expect(wr.Write([]byte{0x01})).To(Equal(1), "should not limit")
			n, err := wr.Write([]byte{0x02})
			Expect(n).To(Equal(0), "should not have written the byte")
			Expect(errors.Is(err, clserve.ErrBufferFull)).To(BeTrue(), "should have been this error")
			Expect(rec.Body.Len()).To(Equal(0), "should not have flushed anything")
		})

		It("should limit writes when writing past", func() {
			rec := httptest.NewRecorder()
			wr = clserve.NewBufferResponse(rec, 1)
			n, err := wr.Write([]byte{0x01, 0x02})
			Expect(n).To(Equal(0))
			Expect(errors.Is(err, clserve.ErrBufferFull)).To(BeTrue())
			Expect(rec.Body.Len()).To(Equal(0))
			Expect(rec.Body.Len()).To(Equal(0), "should not have flushed anything")
		})

		It("should not limit writes when passed -1", func() {
			rec := httptest.NewRecorder()
			wr = clserve.NewBufferResponse(rec, -1)
			Expect(wr.Write([]byte{0x01, 0x02})).To(Equal(2))
			Expect(rec.Body.Len()).To(Equal(0), "should not have flushed anything")
		})

		It("should flush correctly", func() {
			rec := httptest.NewRecorder()
			fwr = clserve.NewBufferResponse(rec, 2)
			for i := 0; i < 3; i++ {
				Expect(fwr.Write([]byte{0x01, 0x02})).To(Equal(2))
				Expect(fwr.FlushError()).To(Succeed())
			}
			Expect(rec.Body.Bytes()).To(Equal([]byte{0x01, 0x02, 0x01, 0x02, 0x01, 0x02}))
		})
	})

	It("should unwrap correctly", func() {
		rec := httptest.NewRecorder()
		fwr = clserve.NewBufferResponse(rec, 0)
		Expect(fwr.Unwrap()).To(Equal(rec))
	})

	It("should pass on flush errors", func() {
		rec := httptest.NewRecorder()
		wr := failingResponseWriter{rec}
		fwr = clserve.NewBufferResponse(wr, -1)
		fmt.Fprintf(fwr, "foo") // without something in the buffer underlying write won't be triggered
		Expect(fwr.FlushError()).To(MatchError("write fail"))
	})

	Describe("reset behaviour", func() {
		It("should allow re-writing after reset", func() {
			rec := httptest.NewRecorder()
			resp = clserve.NewBufferResponse(rec, -1)

			Expect(fmt.Fprintf(resp, "foo")).To(Equal(3))
			Expect(resp.Reset()).To(Succeed())
			Expect(fmt.Fprintf(resp, "bar")).To(Equal(3))
			Expect(resp.FlushError()).To(Succeed())
			Expect(rec.Body.String()).To(Equal("bar"))
		})

		It("should allow re-writing headers", func() {
			rec := httptest.NewRecorder()
			resp = clserve.NewBufferResponse(rec, -1)
			resp.Header().Set("X-Before", "before")
			Expect(resp.Reset()).To(Succeed())
			resp.Header().Set("X-After", "after")
			Expect(resp.FlushError()).To(Succeed())

			Expect(rec.Header().Get("X-After")).To(Equal("after"))
			Expect(rec.Header().Values("X-Bfore")).ToNot(Equal(nil))
		})

		It("should allow re-writing status code", func() {
			rec := httptest.NewRecorder()
			resp = clserve.NewBufferResponse(rec, -1)
			resp.WriteHeader(201)
			Expect(resp.Reset()).To(Succeed())
			resp.WriteHeader(202)
			Expect(resp.FlushError()).To(Succeed())
			Expect(rec.Code).To(Equal(202))
		})

		It("should reset to default status of 200", func() {
			rec := httptest.NewRecorder()
			resp = clserve.NewBufferResponse(rec, -1)
			resp.WriteHeader(201)
			Expect(resp.Reset()).To(Succeed())
			Expect(resp.FlushError()).To(Succeed())
			Expect(rec.Code).To(Equal(200))
		})

		It("should not allow reset after explicit flush", func() {
			rec := httptest.NewRecorder()
			resp = clserve.NewBufferResponse(rec, -1)
			rc := http.NewResponseController(resp)
			Expect(rc.Flush()).To(Succeed())

			err := resp.Reset()
			Expect(errors.Is(err, clserve.ErrAlreadyFlushed)).To(BeTrue())
		})

		It("should reset limit after reset", func() {
			rec := httptest.NewRecorder()
			resp = clserve.NewBufferResponse(rec, 2)
			for i := 0; i < 3; i++ {
				Expect(resp.Reset()).To(Succeed())
				Expect(resp.Write([]byte("fo"))).To(Equal(2))

			}

			Expect(resp.FlushError()).To(Succeed())
			Expect(rec.Body.String()).To(Equal("fo"))
		})
	})
})
