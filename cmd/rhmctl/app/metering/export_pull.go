package metering

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/util/i18n"
)

type exportPullOptions struct {
}

func (pull *exportPullOptions) Run() error {

	return nil
}

func NewCmdExportPull(conf *rhmctlapi.Config, f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := exportPullOptions{}

	cmd := &cobra.Command{
		Use:                   "pull [(--before DATE) (--after DATE) (--include-deleted)]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Pulls files from RHM Operator"),
		// Long:                  imageLong,
		// Example:               imageExample,
		Run: func(cmd *cobra.Command, args []string) {
			// cmdutil.CheckErr(o.Complete(f, cmd, args))
			// cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
			cmd.Help()
		},
	}

	return cmd
}
