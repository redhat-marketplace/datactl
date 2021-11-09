package clients

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"io/ioutil"
	"strings"

	"emperror.dev/errors"
	"github.com/redhat-marketplace/rhmctl/pkg/clients/dataservice"
	"github.com/redhat-marketplace/rhmctl/pkg/clients/marketplace"
	rhmctlapi "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
)

func ProvideDataService(
	currentContext string,
	rhmRawConfig *rhmctlapi.Config,
) (*dataservice.DataServiceConfig, error) {
	dsConfig, exists := rhmRawConfig.DataServiceEndpoints[currentContext]

	if !exists {
		return nil, errors.New("data-service is not configured, run `rhmctl config init`")
	}

	tlsConfig := &tls.Config{}

	if strings.HasPrefix(dsConfig.URL, "https") {
		if dsConfig.InsecureSkipTLSVerify {
			tlsConfig.InsecureSkipVerify = true
		} else {
			tlsConfig.RootCAs = x509.NewCertPool()

			if dsConfig.CertificateAuthority != "" {
				data, err := ioutil.ReadFile(dsConfig.CertificateAuthority)
				if err != nil {
					return nil, errors.WithMessage(err, "failed to read certificate authority file data from rhmctl config")
				}
				ok := tlsConfig.RootCAs.AppendCertsFromPEM(data)

				if !ok {
					return nil, errors.New("failed to read certificate authority file data from rhmctl config")
				}
			} else if len(dsConfig.CertificateAuthorityData) != 0 {
				data := []byte{}
				_, err := base64.StdEncoding.Decode(data, dsConfig.CertificateAuthorityData)

				if err != nil {
					return nil, errors.WithMessage(err, "failed to read certificate authority data from rhmctl config")
				}

				ok := tlsConfig.RootCAs.AppendCertsFromPEM(data)
				if !ok {
					return nil, errors.New("failed to read certificate authority data from rhmctl config")
				}
			}
		}
	}

	var token string
	if dsConfig.Token != "" {
		data, err := ioutil.ReadFile(dsConfig.Token)
		if err != nil {
			return nil, err
		}
		token = string(data)
	}

	if dsConfig.TokenData != "" {
		data, err := base64.StdEncoding.DecodeString(dsConfig.TokenData)
		if err != nil {
			return nil, err
		}
		token = string(data)
	}

	return &dataservice.DataServiceConfig{
		URL:       dsConfig.URL,
		Token:     token,
		TlsConfig: tlsConfig,
	}, nil
}

func ProvideMarketplaceUpload(
	rhmRawConfig *rhmctlapi.Config,
) (*marketplace.MarketplaceConfig, error) {
	return nil, nil
}
