package metering

import (
	rhmctlapi "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
)

// https://github.com/gofrs/flock

func NewCmdExportStart(conf *rhmctlapi.Config, f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := exportStartOptions{
		configFlags: genericclioptions.NewConfigFlags(false),
	}

	cmd := &cobra.Command{
		Use:                   "s [(-n|--name)=NAME]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Starts an export from the RHM Operator"),
		// Long:                  imageLong,
		// Example:               imageExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}

	cmd.Flags().StringVarP(&o.name, "name", "n", "", i18n.T("name of the file (default is to generate one)"))
	o.configFlags.AddFlags(cmd.Flags())

	return cmd
}

type exportStartOptions struct {
	configFlags *genericclioptions.ConfigFlags

	name string

	args      []string
	rawConfig clientapi.Config
}

func (s *exportStartOptions) Complete(cmd *cobra.Command, args []string) error {
	s.args = args

	var err error
	s.rawConfig, err = s.configFlags.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return err
	}

	return nil
}

func (s *exportStartOptions) Validate() error {
	return nil
}

func (s *exportStartOptions) Run() error {

	return nil
}
