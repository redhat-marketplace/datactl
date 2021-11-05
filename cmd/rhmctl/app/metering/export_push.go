package metering

import (
	rhmctlapi "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
)

type exportPushOptions struct {
}

func (push *exportPushOptions) Run() error {

	return nil
}

func NewCmdExportPush(conf *rhmctlapi.Config, f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := exportCommitOptions{}

	cmd := &cobra.Command{
		Use:                   "push",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Pushes downloaded files."),
		//Long:                  imageLong,
		//Example:               imageExample,
		Run: func(cmd *cobra.Command, args []string) {
			// cmdutil.CheckErr(o.Complete(f, cmd, args))
			// cmdutil.CheckErr(o.Validate())
			cmd.Help()
			cmdutil.CheckErr(o.Run())
		},
	}

	return cmd
}
