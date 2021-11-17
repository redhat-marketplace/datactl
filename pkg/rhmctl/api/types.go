package api

import (
	dataservicev1 "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api/dataservice/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Config struct {
	MarketplaceEndpoint Marketplace `json:"marketplace"`

	MeteringExports map[string]*MeteringExport `json:"metering-export-history,omitempty"`

	DataServiceEndpoints map[string]*DataServiceEndpoint `json:"data-service-endpoints"`
}

type MeteringExport struct {
	// LocationOfOrigin indicates where this object came from.  It is used for round tripping config post-merge, but never serialized.
	// +k8s:conversion-gen=false
	LocationOfOrigin string `json:"-"`

	FileName string `json:"name"`

	// +optional
	DataServiceContext string `json:"data-service-context,omitempty"`

	// +optional
	Files []*dataservicev1.FileInfoCTLAction `json:"files,omitempty"`

	// +k8s:conversion-gen=false
	Committed bool `json:"-"`

	// +k8s:conversion-gen=false
	Pushed bool `json:"-"`

	// DEPRECATED
	// +optional
	Start *metav1.Time `json:"start,omitempty"`

	// DEPRECATED
	// +optional
	End *metav1.Time `json:"end,omitempty"`

	// DEPRECATED
	// +optional
	Active bool `json:"active,omitempty"`

	// DEPRECATED
	// +optional
	DFileInfo []*MeteringFileSummary `json:"info,omitempty"`
}

// DEPRECATED
type MeteringFileSummary struct {
	DataServiceContext string `json:"data-service-context"`

	// +optional
	Files []*dataservicev1.FileInfoCTLAction `json:"files,omitempty"`

	Committed bool `json:"committed,omitempty"`

	Pushed bool `json:"pushed,omitempty"`
}

type Marketplace struct {
	// LocationOfOrigin indicates where this object came from.  It is used for round tripping config post-merge, but never serialized.
	// +k8s:conversion-gen=false
	LocationOfOrigin string

	// Host is the url of the marketplace i.e. marketplace.redhat.com
	Host string `json:"host"`

	// +optional
	PullSecret string `json:"pull-secret,omitempty"`

	// +optional
	PullSecretData string `json:"pull-secret-data,omitempty"`

	// InsecureSkipTLSVerify skips the validity check for the server's certificate. This will make your HTTPS connections insecure.
	// +optional
	InsecureSkipTLSVerify bool `json:"insecure-skip-tls-verify,omitempty"`

	// CertificateAuthority is the path to a cert file for the certificate authority.
	// +optional
	CertificateAuthority string `json:"certificate-authority,omitempty"`

	// CertificateAuthorityData contains PEM-encoded certificate authority certificates. Overrides CertificateAuthority
	// +optional
	CertificateAuthorityData []byte `json:"certificate-authority-data,omitempty"`

	// ProxyURL is the URL to the proxy to be used for all requests made by this
	// client. URLs with "http", "https", and "socks5" schemes are supported.  If
	// this configuration is not provided or the empty string, the client
	// attempts to construct a proxy configuration from http_proxy and
	// https_proxy environment variables. If these environment variables are not
	// set, the client does not attempt to proxy requests.
	//
	// socks5 proxying does not currently support spdy streaming endpoints (exec,
	// attach, port forward).
	// +optional
	ProxyURL string `json:"proxy-url,omitempty"`
}

type DataServiceEndpoint struct {
	// LocationOfOrigin indicates where this object came from.  It is used for round tripping config post-merge, but never serialized.
	// +k8s:conversion-gen=false
	LocationOfOrigin string

	ClusterContextName string `json:"cluster-context-name"`

	URL string `json:"url"`

	// Token is a filepath to a token file
	Token string `json:"token,omitempty"`

	// TokenData is base64 encoded token in the config file, env var, or token argument
	TokenData string `json:"token-data,omitempty"`

	ServiceAccount string `json:"service-account,omitempty"`

	// InsecureSkipTLSVerify skips the validity check for the server's certificate. This will make your HTTPS connections insecure.
	// +optional
	InsecureSkipTLSVerify bool `json:"insecure-skip-tls-verify,omitempty"`

	// CertificateAuthority is the path to a cert file for the certificate authority.
	// +optional
	CertificateAuthority string `json:"certificate-authority,omitempty"`

	// CertificateAuthorityData contains PEM-encoded certificate authority certificates. Overrides CertificateAuthority
	// +optional
	CertificateAuthorityData []byte `json:"certificate-authority-data,omitempty"`

	// ProxyURL is the URL to the proxy to be used for all requests made by this
	// client. URLs with "http", "https", and "socks5" schemes are supported.  If
	// this configuration is not provided or the empty string, the client
	// attempts to construct a proxy configuration from http_proxy and
	// https_proxy environment variables. If these environment variables are not
	// set, the client does not attempt to proxy requests.
	//
	// socks5 proxying does not currently support spdy streaming endpoints (exec,
	// attach, port forward).
	// +optional
	ProxyURL string `json:"proxy-url,omitempty"`
}

func NewConfig() *Config {
	return &Config{
		DataServiceEndpoints: make(map[string]*DataServiceEndpoint),
		MarketplaceEndpoint:  Marketplace{},
		MeteringExports:      make(map[string]*MeteringExport),
	}
}

const (
	marketplaceProductionUrl = "https://marketplace.redhat.come"
)

func NewDefaultConfig() *Config {
	conf := NewConfig()
	conf.MarketplaceEndpoint.Host = marketplaceProductionUrl
	return conf
}

func NewDefaultMeteringExport() *MeteringExport {
	export := MeteringExport{}
	export.Start = nil
	export.End = nil
	return &export
}
