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

package v1

import (
	api "github.com/redhat-marketplace/datactl/pkg/datactl/api"
	dataservicev1 "github.com/redhat-marketplace/datactl/pkg/datactl/api/dataservice/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Config struct {
	MarketplaceEndpoint UploadAPI `json:"upload-api"`

	CurrentMeteringExport *MeteringExport `json:"current-metering-export,omitempty"`

	MeteringExports []*MeteringExport `json:"metering-export-history,omitempty"`

	DataServiceEndpoints []*DataServiceEndpoint `json:"data-service-endpoints,omitempty"`

	LMTEndpoints []*LMTEndpoint `json:"lmt-endpoints,omitempty"`

	Sources []*Source `json:"sources,omitempty"`
}

type MeteringExport struct {
	FileName string `json:"name"`

	// +optional
	DataServiceCluster string `json:"data-service-cluster,omitempty"`

	// +optional
	Files []*dataservicev1.FileInfoCTLAction `json:"files,omitempty"`
}

type Source struct {
	Name string `json:"source-name"`

	Type api.SourceType `json:"source-type"`

	LastAccessTime metav1.Time `json:"last-access-time,omitempty"`
}

func (s *Source) String() string {
	return s.Type.String() + ":" + s.Name
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

type LMTEndpoint struct {
	// LocationOfOrigin indicates where this object came from.  It is used for round tripping config post-merge, but never serialized.
	// +k8s:conversion-gen=false
	LocationOfOrigin string

	LMTSourceType string `json:"lmtSourceType"`

	Host string `json:"host"`

	Port int `json:"port"`

	// Token is base64 encoded token in the config file, env var, or token argument
	Token string `json:"token"`
}
