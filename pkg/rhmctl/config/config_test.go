package config

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("config", func() {
	var (
		testConfig = `
marketplace:
  host: test.com
data-service-endpoints:
  - cluster-context-name: my-context
    url: "https://foo.test"
    insecure-skip-tls-verify: true`

		f    *os.File
		name string
	)

	BeforeEach(func() {
		var err error
		tmpDir := os.TempDir()

		f, err = ioutil.TempFile(tmpDir, "config*.yaml")
		Expect(err).To(Succeed())

		_, err = f.WriteString(testConfig)
		Expect(err).To(Succeed())

		name = f.Name()

		Expect(f.Close()).To(Succeed())
	})

	AfterEach(func() {
		if f != nil {
			os.Remove(name)
		}
	})

	It("should parse v1", func() {
		conf, err := LoadFromFile(name)
		Expect(err).To(Succeed())
		Expect(conf).ToNot(BeNil())
		Expect(conf.DataServiceEndpoints).To(HaveLen(1))
		Expect(conf.DataServiceEndpoints["my-context"].URL).To(Equal("https://foo.test"))
	})
})
