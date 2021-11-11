package config

import (
	rhmctlapi "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
)

func LatestContextMeteringFileSummary(export *rhmctlapi.MeteringExport, currentContext string) *rhmctlapi.MeteringFileSummary {
	var exportInfo *rhmctlapi.MeteringFileSummary
	for i := range export.FileInfo {
		info := export.FileInfo[i]
		if info.DataServiceContext == currentContext {
			exportInfo = info
		}
	}

	if exportInfo == nil {
		exportInfo = &rhmctlapi.MeteringFileSummary{}
		exportInfo.DataServiceContext = currentContext
		exportInfo.Committed = false

		export.FileInfo = append(export.FileInfo, exportInfo)
	}

	return exportInfo
}
