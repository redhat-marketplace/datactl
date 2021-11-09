package config

import (
	"emperror.dev/errors"
	rhmctlapi "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	"github.com/redhat-marketplace/rhmctl/pkg/rhmctl/config"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
)

func NewCmdConfigInit(rhmFlags *config.ConfigFlags, f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := configInitOptions{
		configFlags:    genericclioptions.NewConfigFlags(false),
		rhmConfigFlags: rhmFlags,
	}

	cmd := &cobra.Command{
		Use:                   "init",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Pulls files from RHM Operator"),
		// Long:                  imageLong,
		// Example:               imageExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}

	return cmd
}

type configInitOptions struct {
	configFlags     *genericclioptions.ConfigFlags
	rhmConfigFlags  *config.ConfigFlags
	rhmConfigAccess config.ConfigAccess

	args         []string
	rhmRawConfig *rhmctlapi.Config
}

func (init *configInitOptions) Complete(cmd *cobra.Command, args []string) error {
	init.args = args

	var err error
	init.rhmRawConfig, err = init.rhmConfigFlags.RawPersistentConfigLoader().RawConfig()
	if err != nil {
		return errors.Wrap(err, "error getting rhm config")
	}

	init.rhmConfigAccess = init.rhmConfigFlags.ConfigAccess()
	return nil
}

func (init *configInitOptions) Validate() error {
	return nil
}

func (init *configInitOptions) Run() error {
	if err := config.ModifyConfig(init.rhmConfigAccess, *init.rhmRawConfig, true); err != nil {
		return errors.Wrap(err, "error modifying config")
	}

	return nil
}
