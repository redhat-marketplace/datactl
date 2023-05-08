package add

import (
	"crypto/x509"
	"strings"

	"github.com/gotidy/ptr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/redhat-marketplace/datactl/pkg/datactl/api"
	"github.com/redhat-marketplace/datactl/pkg/datactl/config"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var _ = Describe("config init", func() {

	It("should handle self signed certificates", func() {
		server := ghttp.NewTLSServer()

		url := server.URL()
		url = strings.TrimPrefix(url, "https://")

		init := &addDataServiceOptions{}

		init.rhmConfigFlags = &config.ConfigFlags{}

		init.rhmConfigFlags.KubectlConfig = &genericclioptions.ConfigFlags{
			Insecure: ptr.Bool(true),
		}

		init.dataServiceConfig = &api.DataServiceEndpoint{
			Host: url,
		}

		err := init.discoverDataServiceCA()
		Expect(err).To(HaveOccurred())

		init.allowSelfsigned = true
		err = init.discoverDataServiceCA()
		Expect(err).To(Succeed())
		Expect(init.dataServiceConfig.CertificateAuthorityData).ToNot(BeEmpty())

		caCertPool, _ := x509.SystemCertPool()
		cert, err := x509.ParseCertificate(init.dataServiceConfig.CertificateAuthorityData)
		Expect(err).To(Succeed())
		caCertPool.AddCert(cert)
	})
})
