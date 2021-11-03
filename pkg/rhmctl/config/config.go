package config

import (
	"os"
	"path/filepath"

	"emperror.dev/errors"
	clientcmdapi "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	clientcmdlatest "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api/latest"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/homedir"
)

var (
	FileName         = "config"
	DirectoryName    = ".rhmctl"
	DefaultDirectory = homedir.HomeDir()
	DefaultPath      = filepath.Join(DefaultDirectory, DirectoryName, FileName)
)

type LoadingRules interface {
	Load() ([]byte, error)
}

type DefaultLoadingRules struct {
	OverridePath string
}

func (l *DefaultLoadingRules) Load() ([]byte, error) {
	if l.OverridePath != "" {
		data, err := os.ReadFile(l.OverridePath)
		if err == nil {
			return data, nil
		}

		logrus.Warnf("failed to read file: %s", err)
		logrus.Warnf("falling back to default config %s", DefaultDirectory)
	}

	data, err := os.ReadFile(DefaultPath)
	if err != nil {
		return data, errors.Wrap(err, "failed to load config")
	}
	return data, nil
}

type LoadingRulesFunc func() ([]byte, error)

func (l LoadingRulesFunc) Load() ([]byte, error) {
	return l()
}

func LoadConfig(l LoadingRules) (*clientcmdapi.Config, error) {
	data, err := l.Load()

	if err != nil {
		return nil, err
	}

	config := clientcmdapi.NewConfig()
	// if there's no data in a file, return the default object instead of failing (DecodeInto reject empty input)
	if len(data) == 0 {
		return config, nil
	}

	decoded, _, err := clientcmdlatest.Codec.Decode(data, &schema.GroupVersionKind{Version: clientcmdlatest.Version, Group: clientcmdlatest.Group, Kind: "Config"}, config)
	if err != nil {
		return nil, err
	}

	return decoded.(*clientcmdapi.Config), nil
}
