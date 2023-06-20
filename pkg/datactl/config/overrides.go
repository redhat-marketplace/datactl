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
	datactlapi "github.com/redhat-marketplace/datactl/pkg/datactl/api"
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/clientcmd"
)

//TODO: add all overrides

type ConfigOverrides struct {
	Marketplace datactlapi.UploadAPI

	CurrentContext string
	Timeout        string

	// TLS
	MinVersion   string
	CipherSuites []string
}

type ConfigOverrideFlags struct {
	Marketplace MarketplaceOverrideFlags

	CurrentContext clientcmd.FlagInfo
	Timeout        clientcmd.FlagInfo

	// TLS
	MinVersion   clientcmd.FlagInfo
	CipherSuites clientcmd.FlagInfo
}

type MarketplaceOverrideFlags struct {
	Host clientcmd.FlagInfo
}

const (
	FlagMarketplaceHost = "upload-api-host"
	FlagContext         = "context"
	FlagTimeout         = "request-timeout"
)

func flagInfo(longName, shortName, defaultVal, description string) clientcmd.FlagInfo {
	return clientcmd.FlagInfo{
		LongName:    longName,
		ShortName:   shortName,
		Default:     defaultVal,
		Description: description,
	}
}

func RecommendedConfigOverrideFlags(prefix string) ConfigOverrideFlags {
	return ConfigOverrideFlags{
		Marketplace: RecommendMarketplaceOverrideFlags(prefix),

		CurrentContext: flagInfo(prefix+FlagContext, "", "", "The name of the kubeconfig context to use"),
		Timeout:        flagInfo(prefix+FlagTimeout, "", "0", "The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests."),
	}
}

func RecommendMarketplaceOverrideFlags(prefix string) MarketplaceOverrideFlags {
	return MarketplaceOverrideFlags{
		Host: flagInfo(prefix+FlagMarketplaceHost, "", "", "Override the Marketplace API host"),
	}
}

func BindOverrideFlags(overrides *ConfigOverrides, flags *pflag.FlagSet, flagNames ConfigOverrideFlags) {
	BindMarketplaceFlags(&overrides.Marketplace, flags, flagNames.Marketplace)
}

func BindMarketplaceFlags(clusterInfo *datactlapi.UploadAPI, flags *pflag.FlagSet, flagNames MarketplaceOverrideFlags) {
	flagNames.Host.BindStringFlag(flags, &clusterInfo.Host)
}
