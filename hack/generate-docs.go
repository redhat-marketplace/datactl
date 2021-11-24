package main

import (
	"log"
	"os"

	"github.com/redhat-marketplace/datactl/cmd/datactl/app"
	"github.com/spf13/cobra/doc"
)

func init() {
	os.Setenv("HOME", "$HOME")
}

func main() {
	folder := os.Args[1]
	cmd := app.NewDefaultDatactlCommand()

	err := doc.GenMarkdownTree(cmd, folder)
	if err != nil {
		log.Fatal(err)
	}
}
