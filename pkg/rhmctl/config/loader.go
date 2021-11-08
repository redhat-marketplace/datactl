package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/imdario/mergo"
	clientcmdapi "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	rhmctlapi "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	"github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api/latest"
	clientcmdlatest "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api/latest"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"
)

type ClientConfigLoader interface {
	// Load returns the latest config
	Load() (*clientcmdapi.Config, error)
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

	return &ClientConfigLoadingRules{
		Precedence:       deduplicate(chain),
		WarnIfAllMissing: warnIfAllMissing,
	}
}

type ClientConfigLoadingRules struct {
	ExplicitPath string
	Precedence   []string

	DoNotResolvePaths bool

	WarnIfAllMissing bool
}

func (rules *ClientConfigLoadingRules) Load() (*clientcmdapi.Config, error) {
	errlist := []error{}
	missingList := []string{}

	rhmctlConfigFiles := []string{}

	// Make sure a file we were explicitly told to use exists
	if len(rules.ExplicitPath) > 0 {
		if _, err := os.Stat(rules.ExplicitPath); os.IsNotExist(err) {
			return nil, err
		}
		rhmctlConfigFiles = append(rhmctlConfigFiles, rules.ExplicitPath)

	} else {
		rhmctlConfigFiles = append(rhmctlConfigFiles, rules.Precedence...)
	}

	rhmctlconfigs := []*clientcmdapi.Config{}
	// read and cache the config files so that we only look at them once
	for _, filename := range rhmctlConfigFiles {
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

		rhmctlconfigs = append(rhmctlconfigs, config)
	}

	if rules.WarnIfAllMissing && len(missingList) > 0 && len(rhmctlconfigs) == 0 {
		klog.Warningf("Config not found: %s", strings.Join(missingList, ", "))
	}

	// first merge all of our maps
	mapConfig := clientcmdapi.NewConfig()

	for _, rhmctlconfig := range rhmctlconfigs {
		mergo.Merge(mapConfig, rhmctlconfig, mergo.WithOverride)
	}

	// merge all of the struct values in the reverse order so that priority is given correctly
	// errors are not added to the list the second time
	nonMapConfig := clientcmdapi.NewConfig()
	for i := len(rhmctlconfigs) - 1; i >= 0; i-- {
		rhmctlconfig := rhmctlconfigs[i]
		mergo.Merge(nonMapConfig, rhmctlconfig, mergo.WithOverride)
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
	clientConfig := NewNonInteractiveDeferredLoadingClientConfig(rules, &ConfigOverrides{})
	rawConfig, err := clientConfig.RawConfig()
	if os.IsNotExist(err) {
		return clientcmdapi.NewConfig(), nil
	}
	if err != nil {
		return nil, err
	}

	return &rawConfig, nil
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
func (rules *ClientConfigLoadingRules) IsDefaultConfig(config *rhmctlapi.Config) bool {
	return false
}

func LoadFromFile(filename string) (*clientcmdapi.Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	klog.V(6).Infoln("Config loaded from file: ", filename)

	decoded, _, err := clientcmdlatest.Codec.Decode(data, &schema.GroupVersionKind{Version: clientcmdlatest.Version, Group: clientcmdlatest.Group, Kind: "Config"}, config)
	if err != nil {
		return nil, err
	}

	config := decoded.(*clientcmdapi.Config)

	for k, v := range config.DataServiceEndpoints {
		v.LocationOfOrigin = filename
		config.DataServiceEndpoints[k] = v
	}

	for k, v := range config.MeteringExports {
		v.LocationOfOrigin = filename
		config.MeteringExports[k] = v
	}

	config.MarketplaceEndpoint.LocationOfOrigin = filename

	if config.CurrentMeteringExport != nil {
		config.CurrentMeteringExport.LocationOfOrigin = filename
	}

	if config.DataServiceEndpoints == nil {
		config.DataServiceEndpoints = make(map[string]*clientcmdapi.DataServiceEndpoint)
	}

	if config.MeteringExports == nil {
		config.MeteringExports = make(map[string]*clientcmdapi.MeteringExport)
	}

	return config, nil
}

func deduplicate(in []string) []string {
	out := make([]string, len(in))
	seen := make(map[string]interface{})

	for _, str := range in {
		if _, ok := seen[str]; ok {
			continue
		}
		seen[str] = str
		out = append(out, str)
	}

	return out
}

// WriteToFile serializes the config to yaml and writes it out to a file.  If not present, it creates the file with the mode 0600.  If it is present
// it stomps the contents
func WriteToFile(config rhmctlapi.Config, filename string) error {
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
		return err
	}
	return nil
}

// Write serializes the config to yaml.
// Encapsulates serialization without assuming the destination is a file.
func Write(config rhmctlapi.Config) ([]byte, error) {
	return runtime.Encode(latest.Codec, &config)
}
