package v1alpha1

import (
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/util"
)

const Version string = "rhmctl/v1alpha1"

func NewRhmCtlConfig() util.VersionedConfig {
	return new(RhmCtlConfig)
}

type RhmCtlConfig struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`

	ClusterContexts []*ClusterContext `yaml:clusterContexts`
}

type ClusterContext struct {
	// Name of the context available in kubernetes/openshift config file.
	Name string `yaml:"name"'`

	DataServiceEndpoint *DataServiceEndpoint `yaml:"dataServiceEndpoint,omitempty"`
}

type DataServiceEndpoint struct {
	Endpoint string `yaml:"endpoint"`

	ServiceAccount *string `yaml:"serviceAccount,omitempty"`

	CertificateAuthorityData *string `yaml:"certificateAuthorityData,omitempty"`

	Insecure bool `yaml:"insecure,omitempty"`
}

type Certificate struct {
	File  string `yaml:file`
	Value string `yaml:certificate`
}

func (c *RhmCtlConfig) GetVersion() string {
	return Version
}
