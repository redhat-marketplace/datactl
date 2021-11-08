package api

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Config struct {
	MarketplaceEndpoint Marketplace `json:"marketplace"`

	CurrentMeteringExport *MeteringExport `json:"metering-export,omitempty"`

	MeteringExports map[string]*MeteringExport `json:"metering-export-history,omitempty"`

	DataServiceEndpoints map[string]*DataServiceEndpoint `json:"data-service-endpoints"`
}

type MeteringExport struct {
	FileName string           `json:"name"`
	Active   bool             `json:"active"`
	Start    metav1.Timestamp `json:"start"`

	// +optional
	End metav1.Timestamp `json:"end,omitempty"`

	// +optional
	FileInfo []*MeteringFileSummary `json:"info,omitempty"`
}

type MeteringFileSummary struct {
	DataServiceContext string `json:"data-service-context"`

	// +optional
	Files []*FileInfo `json:"files,omitempty"`

	Committed bool `json:"committed"`
}

type Marketplace struct {
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
	ClusterContextName string `json:"cluster-context-name"`

	URL string `json:"url"`

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
	}
}

const (
	marketplaceProductionUrl = "https://marketplace.redhat.come"
)

func NewDefaultConfig() *Config {
	conf := NewConfig()
	conf.MarketplaceEndpoint.Host = marketplaceProductionUrl
	return NewDefaultConfig()
}
