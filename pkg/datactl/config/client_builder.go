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
	"fmt"
	"io"
	"sync"

	"github.com/redhat-marketplace/datactl/pkg/clients"
	"github.com/redhat-marketplace/datactl/pkg/clients/dataservice"
	"github.com/redhat-marketplace/datactl/pkg/clients/ilmt"
	"github.com/redhat-marketplace/datactl/pkg/clients/marketplace"
	"github.com/redhat-marketplace/datactl/pkg/clients/serviceaccount"
	"github.com/redhat-marketplace/datactl/pkg/datactl/api"
	datactlapi "github.com/redhat-marketplace/datactl/pkg/datactl/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	k8sapiflag "k8s.io/component-base/cli/flag"
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

func (config *DeferredLoadingClientConfig) DataServiceClientConfig(source api.Source) (*dataservice.DataServiceConfig, error) {
	mergedClientConfig, err := config.createClientConfig()
	if err != nil {
		return nil, err
	}

	ds, err := mergedClientConfig.DataServiceClientConfig(source)
	if clientcmd.IsEmptyConfig(err) {
		return nil, genericclioptions.ErrEmptyConfig
	}

	return ds, err
}

func (config *DeferredLoadingClientConfig) IlmtClientConfig(source api.Source) (*ilmt.IlmtConfig, error) {
	mergedClientConfig, err := config.createClientConfig()
	if err != nil {
		return nil, err
	}

	ilmt, err := mergedClientConfig.IlmtClientConfig(source)
	if clientcmd.IsEmptyConfig(err) {
		return nil, genericclioptions.ErrEmptyConfig
	}

	return ilmt, err
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

// DirectClientConfig is a ClientConfig interface that is backed by a clientcmdapi.Config,
// options overrides, and an optional fallbackReader for auth information.
// Is responsible for generating the result.
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

	logger.Info("TLS", "MinVersion", config.overrides.MinVersion)
	tlsVersion, err := k8sapiflag.TLSVersion(config.overrides.MinVersion)
	if err != nil {
		logger.Error(err, "TLS version invalid")
		return nil, err
	}

	logger.Info("TLS", "CipherSuites", config.overrides.CipherSuites)
	tlsCipherSuites, err := k8sapiflag.TLSCipherSuites(config.overrides.CipherSuites)
	if err != nil {
		logger.Error(err, "failed to convert TLS cipher suite name to ID")
		return nil, err
	}

	mktplConfig.TlsConfig.MinVersion = tlsVersion
	mktplConfig.TlsConfig.CipherSuites = tlsCipherSuites

	if err != nil {
		return nil, err
	}

	return mktplConfig, nil
}

func (config *DirectClientConfig) DataServiceClientConfig(source api.Source) (*dataservice.DataServiceConfig, error) {
	datactlConfig, err := config.RawConfig()
	if err != nil {
		return nil, err
	}

	dsConfig, exists := datactlConfig.DataServiceEndpoints[source.Name]

	if !exists {
		return nil, fmt.Errorf("data-service with name %s not found", source.Name)
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
			dsConfig.Namespace = "redhat-marketplace"
		}

		sa := serviceaccount.NewServiceAccountClient(dsConfig.Namespace, client)

		// TODO make this paramaterized
		token, expires, err := sa.NewServiceAccountToken(dsConfig.ServiceAccount, fmt.Sprintf("rhm-data-service.%s.svc", dsConfig.Namespace), 3600)
		if err != nil || token == "" {
			logger.Info("failed to get service account token", "err", err)
			return nil, err
		}

		dsConfig.TokenData = token
		dsConfig.TokenExpiration = expires
	}

	ds, err := clients.ProvideDataService(dsConfig)

	if err != nil {
		logger.Info("failed to get dataservice", "err", err)
		return nil, err
	}

	logger.Info("TLS", "MinVersion", config.overrides.MinVersion)
	tlsVersion, err := k8sapiflag.TLSVersion(config.overrides.MinVersion)
	if err != nil {
		logger.Error(err, "TLS version invalid")
		return nil, err
	}

	logger.Info("TLS", "CipherSuites", config.overrides.CipherSuites)
	tlsCipherSuites, err := k8sapiflag.TLSCipherSuites(config.overrides.CipherSuites)
	if err != nil {
		logger.Error(err, "failed to convert TLS cipher suite name to ID")
		return nil, err
	}

	ds.TlsConfig.MinVersion = tlsVersion
	ds.TlsConfig.CipherSuites = tlsCipherSuites

	return ds, nil
}

func (config *DirectClientConfig) IlmtClientConfig(source api.Source) (*ilmt.IlmtConfig, error) {
	datactlConfig, err := config.RawConfig()
	if err != nil {
		return nil, err
	}

	ilmtConfig, exists := datactlConfig.ILMTEndpoints[source.Name]

	if !exists {
		return nil, fmt.Errorf("ILMT host with name %s not found", source.Name)
	}

	ilmt, err := clients.ProvideIlmtSource(ilmtConfig)

	if err != nil {
		logger.Info("failed to get ILMT source", "err", err)
		return nil, err
	}

	logger.Info("TLS", "MinVersion", config.overrides.MinVersion)
	tlsVersion, err := k8sapiflag.TLSVersion(config.overrides.MinVersion)
	if err != nil {
		logger.Error(err, "TLS version invalid")
		return nil, err
	}

	logger.Info("TLS", "CipherSuites", config.overrides.CipherSuites)
	tlsCipherSuites, err := k8sapiflag.TLSCipherSuites(config.overrides.CipherSuites)
	if err != nil {
		logger.Error(err, "failed to convert TLS cipher suite name to ID")
		return nil, err
	}

	ilmt.TlsConfig.MinVersion = tlsVersion
	ilmt.TlsConfig.CipherSuites = tlsCipherSuites

	return ilmt, nil
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

	if datactlConfig.CurrentMeteringExport == nil {
		currentMeteringExport = &datactlapi.MeteringExport{}
		datactlConfig.CurrentMeteringExport = currentMeteringExport
		return currentMeteringExport, err
	}

	return datactlConfig.CurrentMeteringExport, err
}
