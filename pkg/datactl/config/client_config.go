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
	"github.com/redhat-marketplace/datactl/pkg/clients/dataservice"
	ilmt "github.com/redhat-marketplace/datactl/pkg/clients/ilmt"
	"github.com/redhat-marketplace/datactl/pkg/clients/marketplace"
	"github.com/redhat-marketplace/datactl/pkg/datactl/api"
	datactlapi "github.com/redhat-marketplace/datactl/pkg/datactl/api"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/tools/clientcmd"
)

type ClientConfig interface {
	// RawConfig returns the merged result of all overrides
	RawConfig() (*datactlapi.Config, error)

	MarketplaceClientConfig() (*marketplace.MarketplaceConfig, error)

	DataServiceClientConfig(source api.Source) (*dataservice.DataServiceConfig, error)

	IlmtClientConfig(source api.Source) (*ilmt.IlmtConfig, error)

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

func (c *clientConfig) DataServiceClientConfig(source api.Source) (*dataservice.DataServiceConfig, error) {
	config, err := c.defaultClientConfig.DataServiceClientConfig(source)
	// replace client-go's ErrEmptyConfig error with our custom, more verbose version
	if clientcmd.IsEmptyConfig(err) {
		return config, genericclioptions.ErrEmptyConfig
	}

	return config, err
}

func (c *clientConfig) IlmtClientConfig(source api.Source) (*ilmt.IlmtConfig, error) {
	config, err := c.defaultClientConfig.IlmtClientConfig(source)

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
