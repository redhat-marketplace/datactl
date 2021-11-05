package metering

import (
	rhmctlapi "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
)

type exportCommitOptions struct {
}

func (commit *exportCommitOptions) Validate(cmd *cobra.Command) error {
	return nil
}

func (commit *exportCommitOptions) Run() error {
	return nil
}

func NewCmdExportCommit(conf *rhmctlapi.Config, f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := exportCommitOptions{}

	cmd := &cobra.Command{
		Use:                   "commit",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Finalizes the download of files."),
		// Long:                  imageLong,
		// Example:               imageExample,
		Run: func(cmd *cobra.Command, args []string) {
			// cmdutil.CheckErr(o.Complete(f, cmd, args))
			// cmdutil.CheckErr(o.Validate())
			cmd.Help()
			cmdutil.CheckErr(o.Run())
		},
	}

	return cmd
}
