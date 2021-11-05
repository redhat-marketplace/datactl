package metering

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/util/i18n"
)


type exportStartOptions struct {
	configFlags *genericclioptions.ConfigFlags
}

func (start *exportStartOptions) Run() error {
	return nil
}

// https://github.com/gofrs/flock

func NewCmdExportStart(conf *rhmctlapi.Config, f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := exportStartOptions{
		configFlags: genericclioptions.NewConfigFlags(false),
	}

	cmd := &cobra.Command{
		Use:                   "start [(-n|--name)=NAME]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Starts an export from the RHM Operator"),
		// Long:                  imageLong,
		// Example:               imageExample,
		Run: func(cmd *cobra.Command, args []string) {
			// cmdutil.CheckErr(o.Complete(f, cmd, args))
			// cmdutil.CheckErr(o.Validate())
			cmd.Help()
			cmdutil.CheckErr(o.Run())
		},
	}

	o.configFlags.

	return cmd
}

