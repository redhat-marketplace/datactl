package config

import (
	"io/ioutil"
	"os"

	"github.com/gotidy/ptr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var _ = Describe("config", func() {
	var (
		testConfig = `
marketplace:
  host: test.com
data-service-endpoints:
  - cluster-name: foo.test
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
		Expect(conf.DataServiceEndpoints["foo.test"]).ToNot(BeNil())
		Expect(conf.DataServiceEndpoints["foo.test"].URL).To(Equal("https://foo.test"))
	})

	It("should read file from flags", func() {
		testFlags := genericclioptions.NewConfigFlags(false)
		testFlags.Context = ptr.String("my-context")
		testFlags.ClusterName = ptr.String("foo")

		rhmConfigFlags := NewConfigFlags(testFlags)
		rhmConfigFlags.RHMCTLConfig = ptr.String(name)

		conf, err := rhmConfigFlags.RawPersistentConfigLoader().RawConfig()
		Expect(err).To(Succeed())
		Expect(conf.DataServiceEndpoints).To(HaveLen(1))
		Expect(conf.DataServiceEndpoints["foo.test"]).ToNot(BeNil())
		Expect(conf.DataServiceEndpoints["foo.test"].URL).To(Equal("https://foo.test"))
	})

	It("should update file", func() {
		testFlags := genericclioptions.NewConfigFlags(false)
		testFlags.ClusterName = ptr.String("foo")
		testFlags.Context = ptr.String("my-context")

		rhmConfigFlags := NewConfigFlags(testFlags)
		rhmConfigFlags.RHMCTLConfig = ptr.String(name)

		conf, err := rhmConfigFlags.RawPersistentConfigLoader().RawConfig()
		Expect(err).To(Succeed())

		conf.MeteringExports["foo"] = &api.MeteringExport{
			FileName:           "foo",
			DataServiceCluster: "foo.test",
		}

		Expect(ModifyConfig(rhmConfigFlags.ConfigAccess(), *conf, true)).To(Succeed())

		conf, err = LoadFromFile(name)
		Expect(err).To(Succeed())

		Expect(conf.DataServiceEndpoints).To(HaveLen(1))
		Expect(conf.DataServiceEndpoints["foo.test"]).ToNot(BeNil())
		Expect(conf.DataServiceEndpoints["foo.test"].URL).To(Equal("https://foo.test"))
		Expect(conf.MeteringExports).To(HaveLen(1))
	})
})
