package clients

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/redhat-marketplace/rhmctl/pkg/clients/dataservice"
	"github.com/redhat-marketplace/rhmctl/pkg/clients/marketplace"
	rhmctlapi "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	"k8s.io/apimachinery/pkg/util/errors"
)

func ProvideDataService(
	dsConfig *rhmctlapi.DataServiceEndpoint,
) (*dataservice.DataServiceConfig, error) {
	errs := []error{}
	tlsConfig := &tls.Config{}

	err := func() error {
		if dsConfig.InsecureSkipTLSVerify {
			tlsConfig.InsecureSkipVerify = true
			return nil
		}

		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}
		tlsConfig.RootCAs = rootCAs

		if dsConfig.CertificateAuthority != "" {
			data, err := ioutil.ReadFile(dsConfig.CertificateAuthority)
			if err != nil {
				return fmt.Errorf("failed to read certificate authority file data from rhmctl config %s", err.Error())
			}
			ok := tlsConfig.RootCAs.AppendCertsFromPEM(data)
			if !ok {
				return fmt.Errorf("failed to append certificate authority file data from rhmctl config")
			}
		} else if len(dsConfig.CertificateAuthorityData) != 0 {
			ok := tlsConfig.RootCAs.AppendCertsFromPEM(dsConfig.CertificateAuthorityData)
			if !ok {
				return fmt.Errorf("failed to read certificate authority data from rhmctl config")
			}
		}

		return nil
	}()

	if err != nil {
		errs = append(errs, err)
	}

	var token string

	err = func() error {
		if dsConfig.TokenData == "" {
			return nil
		}

		data, err := base64.StdEncoding.DecodeString(dsConfig.TokenData)
		if err != nil {
			token = dsConfig.TokenData
			return nil
		}

		token = string(data)
		return nil
	}()

	if err != nil {
		errs = append(errs, err)
	}

	token = strings.TrimSpace(token)

	if len(errs) != 0 {
		return nil, errors.NewAggregate(errs)
	}

	if token == "" {
		return nil, fmt.Errorf("token or token-data not provided")
	}

	url := dsConfig.Host

	if !strings.HasPrefix(url, "https://") {
		url = fmt.Sprintf("https://" + url)
	}

	return &dataservice.DataServiceConfig{
		URL:       url,
		Token:     token,
		TlsConfig: tlsConfig,
	}, nil
}

func ProvideMarketplaceUpload(
	rhmRawConfig *rhmctlapi.Config,
) (*marketplace.MarketplaceConfig, error) {
	mktplConfig := rhmRawConfig.MarketplaceEndpoint

	var token string

	if mktplConfig.PullSecret != "" {
		data, err := ioutil.ReadFile(mktplConfig.PullSecret)
		if err != nil {
			return nil, err
		}
		token = string(data)
	}

	if mktplConfig.PullSecretData != "" {
		data, err := base64.StdEncoding.DecodeString(mktplConfig.PullSecretData)
		if err != nil {
			token = mktplConfig.PullSecretData
		} else {
			token = string(data)
		}
	}

	token = strings.TrimSpace(token)

	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	return &marketplace.MarketplaceConfig{
		URL:   rhmRawConfig.MarketplaceEndpoint.Host,
		Token: token,
		TlsConfig: &tls.Config{
			RootCAs: rootCAs,
		},
	}, nil
}
