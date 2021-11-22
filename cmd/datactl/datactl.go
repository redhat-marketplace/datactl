package main

import (
	"os"

	"github.com/redhat-marketplace/datactl/cmd/datactl/app"
	"k8s.io/component-base/cli"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	command := app.NewDefaultDatactlCommand()
	code := cli.Run(command)
	os.Exit(code)
}
