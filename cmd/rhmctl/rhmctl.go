package main

import (
	"context"
	"os"

	"github.com/redhat-marketplace/rhmctl/cmd/rhmctl/app"
	"k8s.io/component-base/cli"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	command := app.NewDefaultRhmCtlCommand()
	code := cli.Run(command)
	os.Exit(code)
}
