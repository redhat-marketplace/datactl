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
	"os"
	"path/filepath"
	"reflect"
	"sort"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	clientcmdapi "github.com/redhat-marketplace/datactl/pkg/datactl/api"
	datactlapi "github.com/redhat-marketplace/datactl/pkg/datactl/api"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2/klogr"
)

var (
	logger logr.Logger = klogr.New().V(5).WithName("pkg/config")
)

// ConfigAccess is used by subcommands and methods in this package to load and modify the appropriate config files
type ConfigAccess interface {
	// GetLoadingPrecedence returns the slice of files that should be used for loading and inspecting the config
	GetLoadingPrecedence() []string
	// GetStartingConfig returns the config that subcommands should being operating against.  It may or may not be merged depending on loading rules
	GetStartingConfig() (*datactlapi.Config, error)
	// GetDefaultFilename returns the name of the file you should write into (create if necessary), if you're trying to create a new stanza as opposed to updating an existing one.
	GetDefaultFilename() string
	// IsExplicitFile indicates whether or not this command is interested in exactly one file.  This implementation only ever does that  via a flag, but implementations that handle local, global, and flags may have more
	IsExplicitFile() bool
	// GetExplicitFile returns the particular file this command is operating against.  This implementation only ever has one, but implementations that handle local, global, and flags may have more
	GetExplicitFile() string
}

const (
	RecommendedConfigPathFlag   = "datactl-config"
	RecommendedConfigPathEnvVar = "DATACTL_CONFIG"
	RecommendedHomeDir          = ".datactl"
	RecommendedFileName         = "config"
	RecommendedSchemaName       = "schema"
)

var (
	RecommendedConfigDir  = filepath.Join(homedir.HomeDir(), RecommendedHomeDir)
	RecommendedHomeFile   = filepath.Join(RecommendedConfigDir, RecommendedFileName)
	RecommendedSchemaFile = filepath.Join(RecommendedConfigDir, RecommendedSchemaName)

	RecommendedDataDir = filepath.Join(RecommendedConfigDir, "data")
)

var (
	// UseModifyConfigLock ensures that access to kubeconfig file using ModifyConfig method
	// is being guarded by a lock file.
	// This variable is intentionaly made public so other consumers of this library
	// can modify its default behavior, but be caution when disabling it since
	// this will make your code not threadsafe.
	UseModifyConfigLock = true
)

func ModifyConfig(configAccess ConfigAccess, newConfig datactlapi.Config, relativizePaths bool) error {
	if UseModifyConfigLock {
		possibleSources := configAccess.GetLoadingPrecedence()
		// sort the possible kubeconfig files so we always "lock" in the same order
		// to avoid deadlock (note: this can fail w/ symlinks, but... come on).
		sort.Strings(possibleSources)
		for _, filename := range possibleSources {
			if err := lockFile(filename); err != nil {
				return err
			}

			defer unlockFile(filename)
		}
	}

	startingConfig, err := configAccess.GetStartingConfig()
	if err != nil {
		return err
	}

	for key, endpoint := range newConfig.DataServiceEndpoints {
		startingEndpoint, exists := startingConfig.DataServiceEndpoints[key]

		destinationFile := endpoint.LocationOfOrigin
		if len(destinationFile) == 0 {
			destinationFile = configAccess.GetDefaultFilename()
		}

		if startingEndpoint == nil {
			startingEndpoint = &datactlapi.DataServiceEndpoint{}
		}

		if err := writeConfig(configAccess, func(in *datactlapi.Config) (bool, error) {
			if !reflect.DeepEqual(in, startingEndpoint) || !exists {
				t := *endpoint
				in.DataServiceEndpoints[key] = &t
				in.DataServiceEndpoints[key].LocationOfOrigin = destinationFile

				// if relativizePaths {
				// 	if err := RelativizeEndpointLocalPaths(in.DataServiceEndpoints[key]); err != nil {
				// 		return false, err
				// 	}
				// }

				return true, nil
			}

			return false, nil
		}); err != nil {
			return err
		}
	}

	newExports := map[string]*datactlapi.MeteringExport{}

	for key, export := range newConfig.MeteringExports {
		startingExport, exists := startingConfig.MeteringExports[key]
		destinationFile := export.LocationOfOrigin

		if len(destinationFile) == 0 {
			destinationFile = configAccess.GetDefaultFilename()
		}

		if startingExport == nil {
			startingExport = &datactlapi.MeteringExport{}
		}

		if !reflect.DeepEqual(newExports[key], startingExport) || !exists {
			newExports[key] = export
			newExports[key].LocationOfOrigin = destinationFile
		}
	}

	if len(newExports) != 0 {
		if err := writeConfig(configAccess,
			func(in *datactlapi.Config) (bool, error) {
				in.MeteringExports = newExports

				return true, nil
			}); err != nil {
			return err
		}
	}

	if err := writeConfig(configAccess,
		func(in *datactlapi.Config) (bool, error) {
			if !reflect.DeepEqual(in, startingConfig.MarketplaceEndpoint) {
				in.MarketplaceEndpoint = newConfig.MarketplaceEndpoint
				return true, nil
			}
			return false, nil
		}); err != nil {
		return err
	}

	return nil
}

func getConfigFromFile(filename string) (*datactlapi.Config, error) {
	config, err := LoadFromFile(filename)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if config == nil {
		config = clientcmdapi.NewConfig()
	}
	return config, nil
}

func writeConfig(
	configAccess ConfigAccess,
	mutate func(*datactlapi.Config) (bool, error),
) error {
	if configAccess.IsExplicitFile() {
		file := configAccess.GetExplicitFile()
		currConfig, err := getConfigFromFile(file)
		if err != nil {
			return err
		}

		writeFile, err := mutate(currConfig)
		if !writeFile {
			return nil
		}

		if err != nil {
			return err
		}

		if err := WriteToFile(*currConfig, file); err != nil {
			return err
		}

		return nil
	}

	for _, file := range configAccess.GetLoadingPrecedence() {
		currConfig, err := getConfigFromFile(file)
		if err != nil {
			return err
		}

		writeFile, err := mutate(currConfig)

		if !writeFile {
			return nil
		}

		if err := WriteToFile(*currConfig, file); err != nil {
			return err
		}

		return nil
	}

	return errors.New("no config found to write preferences")
}

func lockName(name string) string {
	return name + ".lock"
}

func lockFile(filename string) error {
	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(lockName(filename), os.O_CREATE|os.O_EXCL, 0)
	if err != nil {
		return errors.Wrap(err, "failed to open lockfile "+lockName(filename))
	}
	f.Close()
	return nil
}

func unlockFile(filename string) error {
	return os.Remove(lockName(filename))
}
