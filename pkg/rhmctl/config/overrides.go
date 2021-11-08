package config

import (
	rhmctlapi "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/clientcmd"
)

//TODO: add all overrides

type ConfigOverrides struct {
	Marketplace rhmctlapi.Marketplace

	CurrentContext string
	Timeout        string
}

type ConfigOverrideFlags struct {
	Marketplace MarketplaceOverrideFlags

	CurrentContext clientcmd.FlagInfo
	Timeout        clientcmd.FlagInfo
}

type MarketplaceOverrideFlags struct {
	Host clientcmd.FlagInfo
}

const (
	FlagMarketplaceHost = "marketplace-host"
	FlagContext         = "context"
	FlagTimeout         = "request-timeout"
)

func RecommendedConfigOverrideFlags(prefix string) ConfigOverrideFlags {
	return ConfigOverrideFlags{
		Marketplace: RecommendMarketplaceOverrideFlags(prefix),

		CurrentContext: clientcmd.FlagInfo{prefix + FlagContext, "", "", "The name of the kubeconfig context to use"},
		Timeout:        clientcmd.FlagInfo{prefix + FlagTimeout, "", "0", "The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests."},
	}
}

func RecommendMarketplaceOverrideFlags(prefix string) MarketplaceOverrideFlags {
	return MarketplaceOverrideFlags{
		Host: clientcmd.FlagInfo{prefix + FlagMarketplaceHost, "", "", "Override the Marketplace API host"},
	}
}

func BindOverrideFlags(overrides *ConfigOverrides, flags *pflag.FlagSet, flagNames ConfigOverrideFlags) {
	BindMarketplaceFlags(&overrides.Marketplace, flags, flagNames.Marketplace)
}

func BindMarketplaceFlags(clusterInfo *rhmctlapi.Marketplace, flags *pflag.FlagSet, flagNames MarketplaceOverrideFlags) {
	flagNames.Host.BindStringFlag(flags, &clusterInfo.Host)
}
