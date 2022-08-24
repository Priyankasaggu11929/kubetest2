package demo_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDemo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Demo Suite")
}

var _ = Describe("Demo", func() {
   var pattern string

   BeforeEach(func() {
       pattern = "sample-test"
   })

   It("matching the patterns", func() {
       Expect(pattern).To(Equal("sample-test"))
    })
})
