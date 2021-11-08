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

	b.MeteringExports = make(map[string]*api.MeteringExport)

	for _, aD := range a.MeteringExports {
		bD := &api.MeteringExport{}
		err := autoConvert_v1_MeteringExport_To_api_MeteringExport(aD, bD, scope)
		if err != nil {
			return err
		}

		if bD.Active && b.CurrentMeteringExport == nil {
			b.CurrentMeteringExport = bD
		}

		b.MeteringExports[bD.FileName] = bD
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

	b.MeteringExports = make([]*MeteringExport, len(a.MeteringExports))

	for _, aD := range a.MeteringExports {
		bD := &MeteringExport{}
		err := autoConvert_api_MeteringExport_To_v1_MeteringExport(aD, bD, scope)
		if err != nil {
			return err
		}
		b.MeteringExports = append(b.MeteringExports, bD)
	}

	return nil
}
