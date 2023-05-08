package sources

import (
	"fmt"
	"path"

	"github.com/go-logr/logr"
	"github.com/redhat-marketplace/datactl/cmd/datactl/app/sources/add"
	"github.com/redhat-marketplace/datactl/pkg/datactl/config"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2/klogr"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	logger logr.Logger = klogr.New().V(5).WithName("sources")
)

func NewCmdSources(rhmFlags *config.ConfigFlags, f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "sources SUBCOMMAND",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Manage datactl sources."),
		Long:                  templates.LongDesc(i18n.T(`The file at `) + path.Join("${HOME}", config.RecommendedHomeDir, config.RecommendedFileName) + i18n.T(` is used for configuration.`)),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(add.NewCmdAdd(rhmFlags, f, streams))
	return cmd
}

func helpErrorf(cmd *cobra.Command, format string, args ...interface{}) error {
	cmd.Help()
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s", msg)
}
