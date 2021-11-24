// Copyright 2021 IBM Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"crypto/x509"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/redhat-marketplace/datactl/pkg/datactl/api"
)

var _ = Describe("config init", func() {
	It("should handle self signed certificates", func() {
		server := ghttp.NewTLSServer()

		url := server.URL()
		url = strings.TrimPrefix(url, "https://")

		init := &configInitOptions{}
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
