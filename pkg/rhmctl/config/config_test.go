package config

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("config", func() {
	var (
		testConfig = `
pull-secret: my-file
data-service-endpoints:
  - cluster-context-name: my-context
    url: "https://foo.test"
    insecure-skip-tls-verify: true`
	)

	It("should parse v1", func() {
		conf, err := LoadConfig(LoadingRulesFunc(func() ([]byte, error) {
			return []byte(testConfig), nil
		}))

		Expect(err).To(Succeed())
		Expect(conf).ToNot(BeNil())
		Expect(conf.DataServiceEndpoints).To(HaveLen(1))
		Expect(conf.DataServiceEndpoints["my-context"].URL).To(Equal("https://foo.test"))
	})
})
