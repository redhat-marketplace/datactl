package add

import (
	"os"
	"path/filepath"

	s "strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redhat-marketplace/datactl/pkg/datactl/config"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var _ = Describe("NewCmdAddIlmt", func() {
	homedir, _ := os.UserHomeDir()
	configPath := filepath.Join(homedir, ".datactl", "config")

	Context("test updating the config to make sure no error coming in saving the source info in config", func() {
		It("success", func() {

			kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()

			o := addIlmtOptions{
				rhmConfigFlags: config.NewConfigFlags(kubeConfigFlags),
				Host:           "ilmtunitte2323ß.ibm.com",
				Port:           "2705",
				Token:          "abcd",
			}

			o.Complete(nil, nil)
			o.Validate()
			err := o.addSourceDtlsToConfig(o.Host, o.Port, o.Token)

			Expect(err).To(Succeed())

			// cleanup
			err = os.Remove(configPath)
			Expect(err).To(Succeed())

		})

	})

	Context("test updating the config to make sure it saves properly", func() {
		It("success", func() {

			kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()

			o := addIlmtOptions{
				rhmConfigFlags: config.NewConfigFlags(kubeConfigFlags),
				Host:           "ilmtunittesting2870.ibm.com",
				Port:           "2705",
				Token:          "xyz",
			}

			o.Complete(nil, nil)
			o.Validate()
			o.addSourceDtlsToConfig(o.Host, o.Port, o.Token)
			config, err := os.ReadFile(configPath)
			configStr := string(config)
			Expect(err).To(Succeed())
			Expect(true).To(Equal(s.Contains(configStr, "source-name: ilmtunittesting2870.ibm.com")))
			Expect(true).To(Equal(s.Contains(configStr, "source-type: ILMT")))

			// cleanup
			err = os.Remove(configPath)
			Expect(err).To(Succeed())
		})

	})

})
