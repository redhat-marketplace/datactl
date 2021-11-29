// Copyright 2021 IBM Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
