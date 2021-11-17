package metering

import (
	"emperror.dev/errors"
	rhmctlapi "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	"github.com/redhat-marketplace/rhmctl/pkg/rhmctl/metering"
)

func createOrUpdateBundle(
	activeContext string,
	rhmRawConfig *rhmctlapi.Config,
) (*rhmctlapi.MeteringExport, *metering.BundleFile, error) {
	var currentMeteringExport *rhmctlapi.MeteringExport

	currentMeteringExport, ok := rhmRawConfig.MeteringExports[activeContext]

	if !ok {
		bundle, err := metering.NewBundleWithDefaultName()
		if err != nil {
			return nil, nil, errors.Wrap(err, "creating bundle")
		}

		currentMeteringExport = &rhmctlapi.MeteringExport{
			FileName: bundle.Name(),
			Active:   true,
			Start:    nil,
		}

		rhmRawConfig.MeteringExports[activeContext] = currentMeteringExport
		return rhmRawConfig.MeteringExports[bundle.Name()], bundle, err
	}

	// setting unused fields to nil
	currentMeteringExport.Start = nil
	currentMeteringExport.End = nil
	//

	bundle, err := metering.NewBundle(currentMeteringExport.FileName)
	return currentMeteringExport, bundle, err
}
