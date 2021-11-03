package v1

import (
	api "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	"k8s.io/apimachinery/pkg/conversion"
)

func Convert_v1_Config_To_api_Config(a *Config, b *api.Config, scope conversion.Scope) error {
	err := autoConvert_v1_Config_To_api_Config(a, b, scope)

	if err != nil {
		return err
	}

	b.DataServiceEndpoints = make(map[string]*api.DataServiceEndpoint)

	for _, aD := range a.DataServiceEndpoints {
		bD := &api.DataServiceEndpoint{}
		err := autoConvert_v1_DataServiceEndpoint_To_api_DataServiceEndpoint(aD, bD, scope)
		if err != nil {
			return err
		}
		b.DataServiceEndpoints[aD.ClusterContextName] = bD
	}

	return nil
}

func Convert_api_Config_To_v1_Config(a *api.Config, b *Config, scope conversion.Scope) error {
	err := autoConvert_api_Config_To_v1_Config(a, b, scope)

	if err != nil {
		return err
	}

	b.DataServiceEndpoints = make([]*DataServiceEndpoint, len(a.DataServiceEndpoints))

	for _, aD := range a.DataServiceEndpoints {
		bD := &DataServiceEndpoint{}
		err := autoConvert_api_DataServiceEndpoint_To_v1_DataServiceEndpoint(aD, bD, scope)
		if err != nil {
			return err
		}
		b.DataServiceEndpoints = append(b.DataServiceEndpoints, bD)
	}

	return nil
}
