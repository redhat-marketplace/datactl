package errors

import (
	"fmt"
)

func ConfigUnknownAPIVersionErr(version string) error {
	return fmt.Errorf("Config provided is an unknown API version %s", version)
}
