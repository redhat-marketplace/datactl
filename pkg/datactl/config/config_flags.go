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

package config

import (
	"sync"

	"github.com/gotidy/ptr"
	"github.com/redhat-marketplace/datactl/pkg/clients/dataservice"
	"github.com/redhat-marketplace/datactl/pkg/clients/marketplace"
	"github.com/redhat-marketplace/datactl/pkg/datactl/api"
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

	dataServiceClient     map[string]dataservice.Client
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

func (f *ConfigFlags) DataServiceClient(source api.Source) (dataservice.Client, error) {
	return f.toPersistentDataServiceClient(source)
}

func (f *ConfigFlags) toPersistentDataServiceClient(source api.Source) (dataservice.Client, error) {
	f.dataServiceClientLock.Lock()
	defer f.dataServiceClientLock.Unlock()

	if f.dataServiceClient == nil {
		f.dataServiceClient = make(map[string]dataservice.Client)
	}

	if c, ok := f.dataServiceClient[source.Name]; ok {
		return c, nil
	}

	config, err := f.RawPersistentConfigLoader().DataServiceClientConfig(source)

	if err != nil {
		return nil, err
	}

	f.dataServiceClient[source.Name] = dataservice.NewClient(config)
	return f.dataServiceClient[source.Name], nil
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
