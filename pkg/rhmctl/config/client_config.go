package config

import (
	"github.com/redhat-marketplace/rhmctl/pkg/clients/dataservice"
	"github.com/redhat-marketplace/rhmctl/pkg/clients/marketplace"
	rhmctlapi "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd"
)

type ClientConfig interface {
	// RawConfig returns the merged result of all overrides
	RawConfig() (*rhmctlapi.Config, error)

	MarketplaceClientConfig() (*marketplace.MarketplaceConfig, error)

	DataServiceClientConfig() (*dataservice.DataServiceConfig, error)
}

type clientConfig struct {
	defaultClientConfig ClientConfig
}

func (c *clientConfig) RawConfig() (*rhmctlapi.Config, error) {
	config, err := c.defaultClientConfig.RawConfig()
	// replace client-go's ErrEmptyConfig error with our custom, more verbose version
	if clientcmd.IsEmptyConfig(err) {
		return config, genericclioptions.ErrEmptyConfig
	}
	return config, err
}

func (c *clientConfig) MarketplaceClientConfig() (*marketplace.MarketplaceConfig, error) {
	config, err := c.defaultClientConfig.MarketplaceClientConfig()
	// replace client-go's ErrEmptyConfig error with our custom, more verbose version
	if clientcmd.IsEmptyConfig(err) {
		return config, genericclioptions.ErrEmptyConfig
	}
	return config, err
}

func (c *clientConfig) DataServiceClientConfig() (*dataservice.DataServiceConfig, error) {
	config, err := c.defaultClientConfig.DataServiceClientConfig()
	// replace client-go's ErrEmptyConfig error with our custom, more verbose version
	if clientcmd.IsEmptyConfig(err) {
		return config, genericclioptions.ErrEmptyConfig
	}

	return config, err
}
