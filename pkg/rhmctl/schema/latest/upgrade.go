package latest

import (
	"errors"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/util"
)

func (c *RhmCtlConfig) Upgrade() (util.VersionedConfig, error) {
	return nil, errors.New("there's no version to upgrade from \"v1alpha1\"")
}
