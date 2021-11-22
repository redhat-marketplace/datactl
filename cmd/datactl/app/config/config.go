package config

import (
	"fmt"
	"path"

	"github.com/redhat-marketplace/datactl/pkg/datactl/config"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

func NewCmdConfig(rhmFlags *config.ConfigFlags, f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	// pathOptions := config.NewDefaultClientConfigLoadingRules()

	// if len(pathOptions.ExplicitFile) == 0 {
	// 	pathOptions.ExplicitFile = config.RecommendedFileName
	// }

	cmd := &cobra.Command{
		Use:                   "config SUBCOMMAND",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Modify datactl files"),
		Long: templates.LongDesc(i18n.T(`
			Modify datactl config files using subcommands like "datactl config set current-context my-context"
			The loading order follows these rules:
			1. If the --`) + config.RecommendedConfigPathFlag + i18n.T(` flag is set, then only that file is loaded. The flag may only be set once and no merging takes place.
			2. If $`) + config.RecommendedConfigPathEnvVar + i18n.T(` environment variable is set, then it is used as a list of paths (normal path delimiting rules for your system). These paths are merged. When a value is modified, it is modified in the file that defines the stanza. When a value is created, it is created in the first file that exists. If no files in the chain exist, then it creates the last file in the list.
			3. Otherwise, `) + path.Join("${HOME}", config.RecommendedConfigDir) + i18n.T(` is used and no merging takes place.`)),
		Run: cmdutil.DefaultSubCommandRun(streams.ErrOut),
	}

	// cmd.PersistentFlags().StringVar(&pathOptions.ExplicitPath, pathOptions.ExplicitFileFlag, pathOptions.ExplicitPath, "use a particular kubeconfig file")
	cmd.AddCommand(NewCmdConfigInit(rhmFlags, f, streams))

	return cmd
}

func helpErrorf(cmd *cobra.Command, format string, args ...interface{}) error {
	cmd.Help()
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s", msg)
}
