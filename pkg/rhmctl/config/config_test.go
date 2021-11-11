package config

import (
	"io/ioutil"
	"os"

	"github.com/gotidy/ptr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	It("should read file from flags", func() {
		rhmConfigFlags := NewConfigFlags("my-context")
		rhmConfigFlags.RHMCTLConfig = ptr.String(name)

		conf, err := rhmConfigFlags.RawPersistentConfigLoader().RawConfig()
		Expect(err).To(Succeed())
		Expect(conf.DataServiceEndpoints).To(HaveLen(1))
		Expect(conf.DataServiceEndpoints["my-context"].URL).To(Equal("https://foo.test"))
	})

	It("should update file", func() {
		rhmConfigFlags := NewConfigFlags("my-context")
		rhmConfigFlags.RHMCTLConfig = ptr.String(name)

		conf, err := rhmConfigFlags.RawPersistentConfigLoader().RawConfig()
		Expect(err).To(Succeed())

		conf.MeteringExports["foo"] = &api.MeteringExport{
			FileName: "foo",
			Start:    metav1.Now(),
		}

		Expect(ModifyConfig(rhmConfigFlags.ConfigAccess(), *conf, true)).To(Succeed())

		conf, err = LoadFromFile(name)
		Expect(err).To(Succeed())

		Expect(conf.DataServiceEndpoints).To(HaveLen(1))
		Expect(conf.DataServiceEndpoints["my-context"].URL).To(Equal("https://foo.test"))
		Expect(conf.MeteringExports).To(HaveLen(1))
	})
})
