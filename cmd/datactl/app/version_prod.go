//go:build prod
// +build prod

package app

import (
	_ "embed"
)

var (
	//go:embed version.txt
	version string
)
