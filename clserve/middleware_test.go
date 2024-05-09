package clserve_test

import (
	"fmt"
	"net/http"

	"github.com/crewlinker/clgo/clserve"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("middleware", func() {
	It("should just return the handler without middleware", func() {
		hdlr1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		hdlr2 := clserve.Use(hdlr1)
		Expect(fmt.Sprint(hdlr1)).To(Equal(fmt.Sprint(hdlr2))) // compare addrs
	})

	It("should wrap in the correct order", func() {
		var res string
		hdlr1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { res += "inner" })
		mw1 := func(n http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				res += "1("
				n.ServeHTTP(w, r)
				res += ")1"
			})
		}
		mw2 := func(n http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				res += "2("
				n.ServeHTTP(w, r)
				res += ")2"
			})
		}
		mw3 := func(n http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				res += "3("
				n.ServeHTTP(w, r)
				res += ")3"
			})
		}

		clserve.Use(hdlr1, mw1, mw2, mw3).ServeHTTP(nil, nil)
		Expect(res).To(Equal("1(2(3(inner)3)2)1"))
	})
})
