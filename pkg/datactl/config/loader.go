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

package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"dario.cat/mergo"
	"emperror.dev/errors"
	clientcmdapi "github.com/redhat-marketplace/datactl/pkg/datactl/api"
	datactlapi "github.com/redhat-marketplace/datactl/pkg/datactl/api"
	"github.com/redhat-marketplace/datactl/pkg/datactl/api/latest"
	clientcmdlatest "github.com/redhat-marketplace/datactl/pkg/datactl/api/latest"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"
)

type ClientConfigLoader interface {
	// Load returns the latest config
	Load() (*clientcmdapi.Config, error)
	ConfigAccess
}

func NewDefaultClientConfigLoadingRules() *ClientConfigLoadingRules {
	chain := []string{}
	warnIfAllMissing := false

	envVarFiles := os.Getenv(RecommendedConfigPathEnvVar)
	if len(envVarFiles) != 0 {
		fileList := filepath.SplitList(envVarFiles)
		// prevent the same path load multiple times
		chain = append(chain, fileList...)
		warnIfAllMissing = true
	} else {
		chain = append(chain, RecommendedHomeFile)
	}

	chain = deduplicate(chain)

	return &ClientConfigLoadingRules{
		Precedence:       chain,
		WarnIfAllMissing: warnIfAllMissing,
	}
}

type ClientConfigLoadingRules struct {
	ExplicitFile      string
	ExplicitPath      string
	Precedence        []string
	DoNotResolvePaths bool
	WarnIfAllMissing  bool
}

func (rules *ClientConfigLoadingRules) Load() (*clientcmdapi.Config, error) {
	errlist := []error{}
	missingList := []string{}

	datactlConfigFiles := []string{}

	// Make sure a file we were explicitly told to use exists
	if len(rules.ExplicitPath) > 0 {
		if _, err := os.Stat(rules.ExplicitPath); os.IsNotExist(err) {
			return nil, err
		}
		datactlConfigFiles = append(datactlConfigFiles, rules.ExplicitPath)

	} else {
		datactlConfigFiles = append(datactlConfigFiles, rules.Precedence...)
	}

	datactlconfigs := []*clientcmdapi.Config{}
	// read and cache the config files so that we only look at them once
	for _, filename := range datactlConfigFiles {
		if len(filename) == 0 {
			// no work to do
			continue
		}

		config, err := LoadFromFile(filename)

		if os.IsNotExist(err) {
			// skip missing files
			// Add to the missing list to produce a warning
			missingList = append(missingList, filename)
			continue
		}

		if err != nil {
			errlist = append(errlist, fmt.Errorf("error loading config file \"%s\": %v", filename, err))
			continue
		}

		datactlconfigs = append(datactlconfigs, config)
	}

	if rules.WarnIfAllMissing && len(missingList) > 0 && len(datactlconfigs) == 0 {
		klog.Warningf("Config not found: %s", strings.Join(missingList, ", "))
	}

	// first merge all of our maps
	mapConfig := clientcmdapi.NewConfig()

	for _, datactlconfig := range datactlconfigs {
		mergo.Merge(mapConfig, datactlconfig, mergo.WithOverride)
	}

	// merge all of the struct values in the reverse order so that priority is given correctly
	// errors are not added to the list the second time
	nonMapConfig := clientcmdapi.NewConfig()
	for i := len(datactlconfigs) - 1; i >= 0; i-- {
		datactlconfig := datactlconfigs[i]
		mergo.Merge(nonMapConfig, datactlconfig, mergo.WithOverride)
	}

	// since values are overwritten, but maps values are not, we can merge the non-map config on top of the map config and
	// get the values we expect.
	config := clientcmdapi.NewConfig()
	mergo.Merge(config, mapConfig, mergo.WithOverride)
	mergo.Merge(config, nonMapConfig, mergo.WithOverride)

	return config, utilerrors.NewAggregate(errlist)
}

// GetLoadingPrecedence implements ConfigAccess
func (rules *ClientConfigLoadingRules) GetLoadingPrecedence() []string {
	if len(rules.ExplicitPath) > 0 {
		return []string{rules.ExplicitPath}
	}

	return rules.Precedence
}

// GetStartingConfig implements ConfigAccess
func (rules *ClientConfigLoadingRules) GetStartingConfig() (*clientcmdapi.Config, error) {
	kubectlConfig := genericclioptions.NewConfigFlags(false)
	clientConfig := NewNonInteractiveDeferredLoadingClientConfig(rules, &ConfigOverrides{}, kubectlConfig)
	rawConfig, err := clientConfig.RawConfig()
	if os.IsNotExist(err) {
		return clientcmdapi.NewDefaultConfig(kubectlConfig)
	}
	if err != nil {
		return nil, err
	}

	return rawConfig, nil
}

// GetDefaultFilename implements ConfigAccess
func (rules *ClientConfigLoadingRules) GetDefaultFilename() string {
	// Explicit file if we have one.
	if rules.IsExplicitFile() {
		return rules.GetExplicitFile()
	}
	// Otherwise, first existing file from precedence.
	for _, filename := range rules.GetLoadingPrecedence() {
		if _, err := os.Stat(filename); err == nil {
			return filename
		}
	}
	// If none exists, use the first from precedence.
	if len(rules.Precedence) > 0 {
		return rules.Precedence[0]
	}
	return ""
}

// IsExplicitFile implements ConfigAccess
func (rules *ClientConfigLoadingRules) IsExplicitFile() bool {
	return len(rules.ExplicitPath) > 0
}

// GetExplicitFile implements ConfigAccess
func (rules *ClientConfigLoadingRules) GetExplicitFile() string {
	return rules.ExplicitPath
}

// IsDefaultConfig returns true if the provided configuration matches the default
func (rules *ClientConfigLoadingRules) IsDefaultConfig(config *datactlapi.Config) bool {
	return false
}

func LoadFromFile(filename string) (*clientcmdapi.Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	logger.Info("config loaded from file", "filename", filename)

	config := clientcmdapi.NewConfig()
	// if there's no data in a file, return the default object instead of failing (DecodeInto reject empty input)
	if len(data) == 0 {
		return config, nil
	}

	decoded, _, err := clientcmdlatest.Codec.Decode(data, &schema.GroupVersionKind{Version: clientcmdlatest.Version, Group: clientcmdlatest.Group, Kind: "Config"}, config)
	if err != nil {
		err = errors.Wrap(err, "failed to decode config")
		return nil, err
	}

	config = decoded.(*clientcmdapi.Config)

	for k, v := range config.DataServiceEndpoints {
		v.LocationOfOrigin = filename
		config.DataServiceEndpoints[k] = v
	}

	for k, v := range config.MeteringExports {
		v.LocationOfOrigin = filename
		config.MeteringExports[k] = v
	}

	config.MarketplaceEndpoint.LocationOfOrigin = filename

	if config.DataServiceEndpoints == nil {
		config.DataServiceEndpoints = make(map[string]*clientcmdapi.DataServiceEndpoint)
	}

	if config.MeteringExports == nil {
		config.MeteringExports = make(map[string]*clientcmdapi.MeteringExport)
	}

	return config, nil
}

func deduplicate(in []string) []string {
	out := []string{}
	seen := make(map[string]interface{})

	for _, str := range in {
		if len(str) == 0 {
			continue
		}
		if _, ok := seen[str]; ok {
			continue
		}

		seen[str] = nil
		out = append(out, str)
	}

	return out
}

// WriteToFile serializes the config to yaml and writes it out to a file.  If not present, it creates the file with the mode 0600.  If it is present
// it stomps the contents
func WriteToFile(config datactlapi.Config, filename string) error {
	content, err := Write(config)
	if err != nil {
		return err
	}
	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(filename, content, 0600); err != nil {
		return errors.Wrap(err, "writing file: "+filename)
	}
	return nil
}

// Write serializes the config to yaml.
// Encapsulates serialization without assuming the destination is a file.
func Write(config datactlapi.Config) ([]byte, error) {
	b := &bytes.Buffer{}
	err := latest.Codec.Encode(&config, b)
	return b.Bytes(), err
}
