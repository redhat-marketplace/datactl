package metering

import (
	"emperror.dev/errors"
	rhmctlapi "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	"github.com/redhat-marketplace/rhmctl/pkg/rhmctl/metering"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createOrUpdateBundle(
	rhmRawConfig *rhmctlapi.Config,
) (*rhmctlapi.MeteringExport, *metering.BundleFile, error) {
	var currentMeteringExport *rhmctlapi.MeteringExport

	for _, export := range rhmRawConfig.MeteringExports {
		if export.Active {
			localExport := export
			currentMeteringExport = localExport
		}
	}

	if currentMeteringExport == nil {
		bundle, err := metering.NewBundleWithDefaultName()
		if err != nil {
			return nil, nil, errors.Wrap(err, "creating bundle")
		}

		currentMeteringExport = &rhmctlapi.MeteringExport{
			FileName: bundle.Name(),
			Active:   true,
			Start:    metav1.Now(),
		}

		rhmRawConfig.MeteringExports[bundle.Name()] = currentMeteringExport
		return rhmRawConfig.MeteringExports[bundle.Name()], bundle, err
	}

	bundle, err := metering.NewBundle(currentMeteringExport.FileName)
	return currentMeteringExport, bundle, err
}
