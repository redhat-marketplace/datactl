package config

import (
	"github.com/redhat-marketplace/datactl/pkg/clients/dataservice"
	"github.com/redhat-marketplace/datactl/pkg/clients/marketplace"
	datactlapi "github.com/redhat-marketplace/datactl/pkg/datactl/api"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd"
)

type ClientConfig interface {
	// RawConfig returns the merged result of all overrides
	RawConfig() (*datactlapi.Config, error)

	MarketplaceClientConfig() (*marketplace.MarketplaceConfig, error)

	DataServiceClientConfig() (*dataservice.DataServiceConfig, error)

	MeteringExport() (*datactlapi.MeteringExport, error)

	ConfigAccess() ConfigAccess
}

type clientConfig struct {
	defaultClientConfig ClientConfig
}

func (c *clientConfig) RawConfig() (*datactlapi.Config, error) {
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

func (c *clientConfig) MeteringExport() (*datactlapi.MeteringExport, error) {
	exp, err := c.defaultClientConfig.MeteringExport()

	if clientcmd.IsEmptyConfig(err) {
		return nil, genericclioptions.ErrEmptyConfig
	}

	return exp, nil
}

func (c *clientConfig) ConfigAccess() ConfigAccess {
	return c.defaultClientConfig.ConfigAccess()
}
