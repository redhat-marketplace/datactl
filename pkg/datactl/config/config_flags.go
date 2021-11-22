package config

import (
	"sync"

	"github.com/gotidy/ptr"
	"github.com/redhat-marketplace/datactl/pkg/clients/dataservice"
	"github.com/redhat-marketplace/datactl/pkg/clients/marketplace"
	datactlapi "github.com/redhat-marketplace/datactl/pkg/datactl/api"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type ConfigFlags struct {
	DATACTLConfig *string
	overrides     *ConfigOverrides

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
	meteringExport     *datactlapi.MeteringExport

	KubectlConfig *genericclioptions.ConfigFlags
}

func NewConfigFlags(kubeFlags *genericclioptions.ConfigFlags) *ConfigFlags {
	return &ConfigFlags{
		overrides:         &ConfigOverrides{},
		DATACTLConfig:     ptr.String(""),
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

	if f.DATACTLConfig != nil {
		loadingRules.ExplicitPath = *f.DATACTLConfig
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

func (f *ConfigFlags) MeteringExport() (*datactlapi.MeteringExport, error) {
	return f.toPersistentMeteringExport()
}

func (f *ConfigFlags) toPersistentMeteringExport() (*datactlapi.MeteringExport, error) {
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
	flags.StringVar(f.DATACTLConfig, "rhm-config", "", "override the rhm config file")
	BindOverrideFlags(f.overrides, flags, RecommendedConfigOverrideFlags("rhm-"))
	f.KubectlConfig.AddFlags(flags)
}
