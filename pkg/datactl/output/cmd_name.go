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

package output

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/redhat-marketplace/datactl/pkg/datactl/config"
	"k8s.io/klog/v2"
)

const (
	shortName = "datactl"
)

func CommandName() string {
	if os.Args[0] == "kubectl" || os.Args[0] == "oc" {
		return os.Args[0] + " " + shortName
	}

	return shortName
}

var (
	recommendedConfigDir  = filepath.Join("$HOME", config.RecommendedHomeDir)
	recommendedHomeFile   = filepath.Join(recommendedConfigDir, config.RecommendedFileName)
	recommendedSchemaFile = filepath.Join(recommendedConfigDir, config.RecommendedSchemaName)

	replaceVals = map[string]interface{}{
		"cmd":               CommandName(),
		"defaultConfigFile": recommendedHomeFile,
		"defaultDataPath":   recommendedSchemaFile,
	}
)

func ReplaceCommandStrings(str string) string {
	buffer := &strings.Builder{}
	t := template.Must(template.New("").Parse(str))
	err := t.Execute(buffer, replaceVals)
	if err != nil {
		klog.Fatal(err)
	}
	return buffer.String()
}
