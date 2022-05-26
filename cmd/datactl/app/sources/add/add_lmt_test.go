package add

import (
	"os"

	s "strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-marketplace/datactl/pkg/datactl/config"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var _ = Describe("NewCmdAddLmt", func() {
	Context("test updating the config to make sure it saves properly", func() {
		It("success", func() {

			kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()

			o := addLmtOptions{
				rhmConfigFlags: config.NewConfigFlags(kubeConfigFlags),
				LmtSourceType:  "ILMT",
				Host:           "ilmtunitte2323ß.ibm.com",
				Port:           2705,
				Token:          "xaxacaxcacbaxaxa",
			}

			o.Complete(nil, nil)
			o.Validate()
			err := o.addSourceDtlsToConfig(o.LmtSourceType, o.Host, o.Port, o.Token)

			Expect(err).To(Succeed())
		})

	})

	Context("test updating the config to make sure it saves properly", func() {
		It("success", func() {

			kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()

			o := addLmtOptions{
				rhmConfigFlags: config.NewConfigFlags(kubeConfigFlags),
				LmtSourceType:  "ILMT",
				Host:           "ilmtunittesting6666777.ibm.com",
				Port:           2705,
				Token:          "vhaxavxahxv",
			}

			o.Complete(nil, nil)
			o.Validate()
			o.addSourceDtlsToConfig(o.LmtSourceType, o.Host, o.Port, o.Token)
			config, err := os.ReadFile("/Users/akhileshsrivastava/.datactl/config")
			configStr := string(config)
			Expect(err).To(Succeed())
			Expect(true).To(Equal(s.Contains(configStr, "source-name: ilmtunittesting6666777.ibm.com")))
			Expect(true).To(Equal(s.Contains(configStr, "source-type: ILMT")))
		})

	})

})
