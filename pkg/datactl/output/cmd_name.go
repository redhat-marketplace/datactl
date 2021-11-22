package output

import (
	"html/template"
	"os"
	"strings"

	"github.com/redhat-marketplace/datactl/pkg/datactl/config"
	"github.com/sirupsen/logrus"
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
	replaceVals = map[string]interface{}{
		"cmd":               CommandName(),
		"defaultConfigFile": config.RecommendedHomeFile,
		"defaultDataPath":   config.RecommendedDataDir,
	}
)

func ReplaceCommandStrings(str string) string {
	buffer := &strings.Builder{}
	t := template.Must(template.New("").Parse(str))
	err := t.Execute(buffer, replaceVals)
	if err != nil {
		logrus.WithError(err).Fatal("failed to parse template")
	}
	return buffer.String()
}
