package config

import (
	"io"
	"sync"

	"github.com/redhat-marketplace/rhmctl/pkg/clients"
	"github.com/redhat-marketplace/rhmctl/pkg/clients/dataservice"
	"github.com/redhat-marketplace/rhmctl/pkg/clients/marketplace"
	rhmctlapi "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd"
)

type DeferredLoadingClientConfig struct {
	loader    ClientConfigLoader
	overrides *ConfigOverrides

	fallbackReader io.Reader
	clientConfig   ClientConfig
	loadingLock    sync.Mutex
}

func (config *DeferredLoadingClientConfig) createClientConfig() (ClientConfig, error) {
	config.loadingLock.Lock()
	defer config.loadingLock.Unlock()

	if config.clientConfig != nil {
		return config.clientConfig, nil
	}

	mergedConfig, err := config.loader.Load()
	if err != nil {
		return nil, err
	}

	var currentContext string
	if config.overrides != nil {
		currentContext = config.overrides.CurrentContext
	}

	config.clientConfig = NewNonInteractiveClientConfig(*mergedConfig, currentContext, config.overrides)
	return config.clientConfig, nil

}

func (config *DeferredLoadingClientConfig) RawConfig() (*rhmctlapi.Config, error) {
	mergedClientConfig, err := config.createClientConfig()
	if err != nil {
		return &rhmctlapi.Config{}, err
	}

	return mergedClientConfig.RawConfig()
}

func (config *DeferredLoadingClientConfig) MarketplaceClientConfig() (*marketplace.MarketplaceConfig, error) {
	mergedClientConfig, err := config.createClientConfig()
	if err != nil {
		return nil, err
	}

	mktpl, err := mergedClientConfig.MarketplaceClientConfig()
	if clientcmd.IsEmptyConfig(err) {
		return nil, genericclioptions.ErrEmptyConfig
	}

	return mktpl, err
}

func (config *DeferredLoadingClientConfig) DataServiceClientConfig() (*dataservice.DataServiceConfig, error) {
	mergedClientConfig, err := config.createClientConfig()
	if err != nil {
		return nil, err
	}

	ds, err := mergedClientConfig.DataServiceClientConfig()
	if clientcmd.IsEmptyConfig(err) {
		return nil, genericclioptions.ErrEmptyConfig
	}

	return ds, err
}

func (config *DeferredLoadingClientConfig) ConfigAccess() ConfigAccess {
	return config.loader
}

func NewNonInteractiveDeferredLoadingClientConfig(loader ClientConfigLoader, overrides *ConfigOverrides) ClientConfig {
	return &DeferredLoadingClientConfig{loader: loader, overrides: overrides, icc: &inClusterClientConfig{overrides: overrides}}
}

func NewNonInteractiveClientConfig(config rhmctlapi.Config, contextName string, overrides *ConfigOverrides, configAccess ConfigAccess) ClientConfig {
	return &DirectClientConfig{config, contextName, overrides, configAccess}
}

// DirectClientConfig is a ClientConfig interface that is backed by a clientcmdapi.Config, options overrides, and an optional fallbackReader for auth information
type DirectClientConfig struct {
	config      rhmctlapi.Config
	contextName string
	overrides   *ConfigOverrides

	ConfigAccess
}

func (config *DirectClientConfig) RawConfig() (*rhmctlapi.Config, error) {
	return &config.config, nil
}

func (config *DirectClientConfig) MarketplaceClientConfig() (*marketplace.MarketplaceConfig, error) {
	mktplConfig, err := clients.ProvideMarketplaceUpload(&config.config)

	if err != nil {
		return nil, err
	}

	return mktplConfig, nil
}

func (config *DirectClientConfig) DataServiceClientConfig() (*dataservice.DataServiceConfig, error) {
	ds, err := clients.ProvideDataService(config.contextName, &config.config)

	if err != nil {
		return nil, err
	}

	return ds, nil
}
