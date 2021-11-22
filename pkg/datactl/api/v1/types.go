package v1

import (
	dataservicev1 "github.com/redhat-marketplace/datactl/pkg/datactl/api/dataservice/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Config struct {
	MarketplaceEndpoint UploadAPI `json:"upload-api"`

	MeteringExports []*MeteringExport `json:"metering-export-history,omitempty"`

	DataServiceEndpoints []*DataServiceEndpoint `json:"data-service-endpoints"`
}

type MeteringExport struct {
	FileName string `json:"name"`

	// +optional
	DataServiceCluster string `json:"data-service-cluster,omitempty"`

	// +optional
	Files []*dataservicev1.FileInfoCTLAction `json:"files,omitempty"`
}

type UploadAPI struct {
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
	ClusterName string `json:"cluster-name"`

	Host string `json:"host"`

	TokenData string `json:"token-data,omitempty"`

	TokenExpiration metav1.Time `json:"token-expiration,omitempty"`

	ServiceAccount string `json:"service-account,omitempty"`

	Namespace string `json:"namespace,omitempty"`

	// InsecureSkipTLSVerify skips the validity check for the server's certificate. This will make your HTTPS connections insecure.
	// +optional
	InsecureSkipTLSVerify bool `json:"insecure-skip-tls-verify,omitempty"`

	// CertificateAuthority is the path to a cert file for the certificate authority.
	// +optional
	CertificateAuthority string `json:"certificate-authority,omitempty"`

	// CertificateAuthorityData contains PEM-encoded certificate authority certificates. Overrides CertificateAuthority
	// +optional
	CertificateAuthorityData []byte `json:"certificate-authority-data,omitempty"`
}
