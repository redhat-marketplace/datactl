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
	ilmt "github.com/redhat-marketplace/datactl/pkg/clients/ilmt"
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

	// TLS config
	MinVersion   *string
	CipherSuites *[]string

	// export flags
	ExportFileName *string

	// dataservice
	DataServiceCAFile *string

	dataServiceClient     map[string]dataservice.Client
	dataServiceClientLock sync.Mutex

	ilmtClient     map[string]ilmt.Client
	ilmtClientLock sync.Mutex

	meteringExportLock sync.Mutex
	meteringExport     *datactlapi.MeteringExport

	KubectlConfig *genericclioptions.ConfigFlags
}

func NewConfigFlags(kubeFlags *genericclioptions.ConfigFlags) *ConfigFlags {
	cipherSuites := &[]string{}
	return &ConfigFlags{
		overrides:         &ConfigOverrides{},
		DATACTLConfig:     ptr.String(""),
		MarketplaceHost:   ptr.String(""),
		MarketplaceToken:  ptr.String(""),
		DataServiceCAFile: ptr.String(""),
		ExportFileName:    ptr.String(""),
		MinVersion:        ptr.String(""),
		CipherSuites:      cipherSuites,
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

	if f.MinVersion != nil {
		f.overrides.MinVersion = *f.MinVersion
	}

	if f.CipherSuites != nil {
		f.overrides.CipherSuites = *f.CipherSuites
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

	f.dataServiceClient[source.Name], err = dataservice.NewClient(config)
	if err != nil {
		return nil, err
	}

	return f.dataServiceClient[source.Name], nil
}

func (f *ConfigFlags) IlmtClient(source api.Source) (ilmt.Client, error) {
	return f.toPersistentIlmtClient(source)
}

func (f *ConfigFlags) toPersistentIlmtClient(source api.Source) (ilmt.Client, error) {
	f.ilmtClientLock.Lock()
	defer f.ilmtClientLock.Unlock()

	if f.ilmtClient == nil {
		f.ilmtClient = make(map[string]ilmt.Client)
	}

	if c, ok := f.ilmtClient[source.Name]; ok {
		return c, nil
	}

	config, err := f.RawPersistentConfigLoader().IlmtClientConfig(source)

	if err != nil {
		return nil, err
	}

	f.ilmtClient[source.Name], err = ilmt.NewClient(config)
	if err != nil {
		return nil, err
	}

	return f.ilmtClient[source.Name], nil
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

	f.marketplaceClient, err = marketplace.NewClient(config)
	if err != nil {
		return nil, err
	}

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
	flags.StringVar(f.MinVersion, "tls-min-version", "VersionTLS12", "Minimum TLS version supported. Value must match version names from https://golang.org/pkg/crypto/tls/#pkg-constants.")
	flags.StringSliceVar(f.CipherSuites,
		"tls-cipher-suites",
		[]string{"TLS_AES_128_GCM_SHA256",
			"TLS_AES_256_GCM_SHA384",
			"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
			"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
			"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
			"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"},
		"Comma-separated list of cipher suites for the server. Values are from tls package constants (https://golang.org/pkg/crypto/tls/#pkg-constants). If omitted, a subset will be used")
	BindOverrideFlags(f.overrides, flags, RecommendedConfigOverrideFlags("rhm-"))
	f.KubectlConfig.AddFlags(flags)
}
