package config

import (
	"sync"

	"github.com/gotidy/ptr"
	"github.com/redhat-marketplace/rhmctl/pkg/clients/dataservice"
	"github.com/redhat-marketplace/rhmctl/pkg/clients/marketplace"
	rhmctlapi "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type ConfigFlags struct {
	RHMCTLConfig *string
	overrides    *ConfigOverrides

	config     ClientConfig
	configLock sync.Mutex

	// config flags
	MarketplaceHost  *string
	MarketplaceToken *string

	marketplaceClient     marketplace.Client
	marketplaceClientLock sync.Mutex

	// export flags
	ExportFileName *string

	// dataservice
	DataServiceCAFile *string

	dataServiceClient     dataservice.Client
	dataServiceClientLock sync.Mutex

	meteringExportLock sync.Mutex
	meteringExport     *rhmctlapi.MeteringExport

	KubectlConfig *genericclioptions.ConfigFlags
}

func NewConfigFlags(kubeFlags *genericclioptions.ConfigFlags) *ConfigFlags {
	return &ConfigFlags{
		overrides:         &ConfigOverrides{},
		RHMCTLConfig:      ptr.String(""),
		MarketplaceHost:   ptr.String(""),
		MarketplaceToken:  ptr.String(""),
		DataServiceCAFile: ptr.String(""),
		ExportFileName:    ptr.String(""),
		KubectlConfig:     kubeFlags,
	}
}

func (f *ConfigFlags) ConfigAccess() ConfigAccess {
	return f.config.ConfigAccess()
}

func (f *ConfigFlags) RawPersistentConfigLoader() ClientConfig {
	return f.toRawPersistentConfigLoader()
}

func (f *ConfigFlags) toRawPersistentConfigLoader() ClientConfig {
	f.configLock.Lock()
	defer f.configLock.Unlock()

	if f.config == nil {
		f.config = f.toRawConfigLoader()
	}

	return f.config
}

func (f *ConfigFlags) toRawConfigLoader() ClientConfig {
	loadingRules := NewDefaultClientConfigLoadingRules()

	if f.RHMCTLConfig != nil {
		loadingRules.ExplicitPath = *f.RHMCTLConfig
	}

	if f.MarketplaceHost != nil {
		f.overrides.Marketplace.Host = *f.MarketplaceHost
	}

	if f.MarketplaceToken != nil {
		f.overrides.Marketplace.PullSecretData = *f.MarketplaceToken
	}

	//TODO add more overrides
	return &clientConfig{
		defaultClientConfig: NewNonInteractiveDeferredLoadingClientConfig(loadingRules, f.overrides, f.KubectlConfig),
	}
}

func (f *ConfigFlags) DataServiceClient() (dataservice.Client, error) {
	return f.toPersistentDataServiceClient()
}

func (f *ConfigFlags) toPersistentDataServiceClient() (dataservice.Client, error) {
	f.dataServiceClientLock.Lock()
	defer f.dataServiceClientLock.Unlock()

	if f.dataServiceClient != nil {
		return f.dataServiceClient, nil
	}

	config, err := f.RawPersistentConfigLoader().DataServiceClientConfig()

	if err != nil {
		return nil, err
	}

	f.dataServiceClient = dataservice.NewClient(config)
	return f.dataServiceClient, nil
}

func (f *ConfigFlags) MarketplaceClient() (marketplace.Client, error) {
	return f.toPersistentMarketplaceClient()
}

func (f *ConfigFlags) toPersistentMarketplaceClient() (marketplace.Client, error) {
	f.marketplaceClientLock.Lock()
	defer f.marketplaceClientLock.Unlock()

	if f.marketplaceClient != nil {
		return f.marketplaceClient, nil
	}

	config, err := f.RawPersistentConfigLoader().MarketplaceClientConfig()
	if err != nil {
		return nil, err
	}

	f.marketplaceClient = marketplace.NewClient(config)
	return f.marketplaceClient, nil
}

func (f *ConfigFlags) MeteringExport() (*rhmctlapi.MeteringExport, error) {
	return f.toPersistentMeteringExport()
}

func (f *ConfigFlags) toPersistentMeteringExport() (*rhmctlapi.MeteringExport, error) {
	f.meteringExportLock.Lock()
	defer f.meteringExportLock.Unlock()

	if f.meteringExport != nil {
		return f.meteringExport, nil
	}

	exp, err := f.RawPersistentConfigLoader().MeteringExport()
	f.meteringExport = exp
	return exp, err
}

func (f *ConfigFlags) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(f.RHMCTLConfig, "rhm-config", "", "override the rhm config file")
	BindOverrideFlags(f.overrides, flags, RecommendedConfigOverrideFlags("rhm-"))
	f.KubectlConfig.AddFlags(flags)
}
