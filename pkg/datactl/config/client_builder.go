package config

import (
	"fmt"
	"io"
	"sync"

	"emperror.dev/errors"
	"github.com/redhat-marketplace/datactl/pkg/clients"
	"github.com/redhat-marketplace/datactl/pkg/clients/dataservice"
	"github.com/redhat-marketplace/datactl/pkg/clients/marketplace"
	"github.com/redhat-marketplace/datactl/pkg/clients/serviceaccount"
	datactlapi "github.com/redhat-marketplace/datactl/pkg/datactl/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type DeferredLoadingClientConfig struct {
	loader    ClientConfigLoader
	overrides *ConfigOverrides

	fallbackReader io.Reader
	clientConfig   ClientConfig
	loadingLock    sync.Mutex
	kubectlConfig  *genericclioptions.ConfigFlags
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

	config.clientConfig = NewNonInteractiveClientConfig(
		*mergedConfig,
		currentContext,
		config.overrides,
		config.loader,
		config.kubectlConfig,
	)
	return config.clientConfig, nil

}

func (config *DeferredLoadingClientConfig) RawConfig() (*datactlapi.Config, error) {
	mergedClientConfig, err := config.createClientConfig()
	if err != nil {
		return &datactlapi.Config{}, err
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
	mergedClientConfig, err := config.createClientConfig()
	if err != nil {
		return nil
	}

	return mergedClientConfig.ConfigAccess()
}

func (config *DeferredLoadingClientConfig) MeteringExport() (*datactlapi.MeteringExport, error) {
	mergedClientConfig, err := config.createClientConfig()
	if err != nil {
		return nil, err
	}

	exp, err := mergedClientConfig.MeteringExport()
	if clientcmd.IsEmptyConfig(err) {
		return nil, genericclioptions.ErrEmptyConfig
	}

	return exp, err
}

func NewNonInteractiveDeferredLoadingClientConfig(loader ClientConfigLoader, overrides *ConfigOverrides, kubectl *genericclioptions.ConfigFlags) ClientConfig {
	return &DeferredLoadingClientConfig{loader: loader, overrides: overrides, kubectlConfig: kubectl}
}

func NewNonInteractiveClientConfig(config datactlapi.Config, contextName string, overrides *ConfigOverrides, configAccess ConfigAccess, kubectl *genericclioptions.ConfigFlags) ClientConfig {
	return &DirectClientConfig{
		config:        config,
		contextName:   contextName,
		overrides:     overrides,
		configAccess:  configAccess,
		kubectlConfig: kubectl,
	}
}

// DirectClientConfig is a ClientConfig interface that is backed by a clientcmdapi.Config, options overrides, and an optional fallbackReader for auth information
type DirectClientConfig struct {
	config        datactlapi.Config
	contextName   string
	overrides     *ConfigOverrides
	configAccess  ConfigAccess
	kubectlConfig *genericclioptions.ConfigFlags
}

func (config *DirectClientConfig) RawConfig() (*datactlapi.Config, error) {
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
	kubeConfig, err := config.kubectlConfig.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return nil, err
	}

	context, ok := kubeConfig.Contexts[kubeConfig.CurrentContext]

	if !ok {
		return nil, errors.New("current kubectl context does now have a configuration")
	}

	datactlConfig, err := config.RawConfig()
	if err != nil {
		return nil, err
	}

	dsConfig, exists := datactlConfig.DataServiceEndpoints[context.Cluster]

	if !exists {
		return nil, fmt.Errorf("data-service is not configured, run %q", "datactl config init")
	}

	restConfig, err := config.kubectlConfig.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	if dsConfig.TokenData == "" || metav1.Now().After(dsConfig.TokenExpiration.Time) {
		if dsConfig.ServiceAccount == "" {
			dsConfig.ServiceAccount = "default"
		}

		if dsConfig.Namespace == "" {
			dsConfig.Namespace = "openshift-redhat-marketplace"
		}

		sa := serviceaccount.NewServiceAccountClient("openshift-redhat-marketplace", client)

		// TODO make this paramaterized
		token, expires, err := sa.NewServiceAccountToken(dsConfig.ServiceAccount, "rhm-data-service.openshift-redhat-marketplace.svc", 3600)
		if err != nil {
			return nil, err
		}

		dsConfig.TokenData = token
		dsConfig.TokenExpiration = expires
	}

	ds, err := clients.ProvideDataService(dsConfig)

	if err != nil {
		return nil, err
	}

	return ds, nil
}

func (config *DirectClientConfig) ConfigAccess() ConfigAccess {
	return config.configAccess
}

func (config *DirectClientConfig) MeteringExport() (*datactlapi.MeteringExport, error) {
	var currentMeteringExport *datactlapi.MeteringExport

	datactlConfig, err := config.RawConfig()
	if err != nil {
		return nil, err
	}

	kubeConfig, err := config.kubectlConfig.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return nil, err
	}

	context, ok := kubeConfig.Contexts[kubeConfig.CurrentContext]

	if !ok {
		return nil, errors.New("current kubectl context does now have a configuration")
	}

	currentMeteringExport, ok = datactlConfig.MeteringExports[context.Cluster]

	if !ok {
		currentMeteringExport = &datactlapi.MeteringExport{
			DataServiceCluster: context.Cluster,
		}

		datactlConfig.MeteringExports[context.Cluster] = currentMeteringExport
		return currentMeteringExport, err
	}

	return currentMeteringExport, err
}
