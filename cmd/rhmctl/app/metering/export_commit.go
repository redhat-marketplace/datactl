package metering

import (
	"context"
	"time"

	"github.com/redhat-marketplace/rhmctl/pkg/clients/dataservice"
	rhmctlapi "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	"github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api/latest"
	"github.com/redhat-marketplace/rhmctl/pkg/rhmctl/config"
	"github.com/redhat-marketplace/rhmctl/pkg/rhmctl/metering"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
)

func NewCmdExportCommit(rhmFlags *config.ConfigFlags, f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := exportCommitOptions{
		configFlags:    genericclioptions.NewConfigFlags(false),
		rhmConfigFlags: rhmFlags,
		PrintFlags:     genericclioptions.NewPrintFlags("export commit").WithTypeSetter(latest.Scheme),
		IOStreams:      ioStreams,
	}

	cmd := &cobra.Command{
		Use:                   "commit [(--dry-run)]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Finalizes the download of files."),
		// Long:                  imageLong,
		// Example:               imageExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}

	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, i18n.T("No action taken. Print only."))
	return cmd
}

type exportCommitOptions struct {
	configFlags    *genericclioptions.ConfigFlags
	rhmConfigFlags *config.ConfigFlags
	PrintFlags     *genericclioptions.PrintFlags

	dryRun bool

	//internal
	args      []string
	rawConfig clientapi.Config

	rhmRawConfig *rhmctlapi.Config
	dataService  dataservice.Client

	currentMeteringExport *rhmctlapi.MeteringExport
	bundle                *metering.BundleFile

	ToPrinter func(string) (printers.ResourcePrinter, error)

	genericclioptions.IOStreams
}

func (c *exportCommitOptions) Complete(cmd *cobra.Command, args []string) error {
	c.args = args

	var err error
	c.rawConfig, err = c.configFlags.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return err
	}

	c.rhmRawConfig, err = c.rhmConfigFlags.RawPersistentConfigLoader().RawConfig()
	if err != nil {
		return err
	}

	c.dataService, err = c.rhmConfigFlags.DataServiceClient()
	if err != nil {
		return err
	}

	c.currentMeteringExport, c.bundle, err = createOrUpdateBundle(c.rhmRawConfig)
	if err != nil {
		return err
	}

	c.ToPrinter = func(operation string) (printers.ResourcePrinter, error) {
		c.PrintFlags.NamePrintFlags.Operation = operation
		return c.PrintFlags.ToPrinter()
	}

	return nil
}

func (c *exportCommitOptions) Validate() error {
	return nil
}

func (c *exportCommitOptions) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	bundle, err := metering.NewBundle(c.currentMeteringExport.FileName)
	if err != nil {
		return err
	}

	defer bundle.Close()

	print, err := c.ToPrinter("commit")
	if err != nil {
		return err
	}

	if c.dryRun {
		logrus.Warn(i18n.T("dry-run enabled, files will not be removed from data service"))
	}

	for _, info := range c.currentMeteringExport.FileInfo {
		for _, file := range info.Files {
			print.PrintObj(file, c.Out)

			if c.dryRun {
				continue
			}

			err := c.dataService.DeleteFile(ctx, file.Id)
			if err != nil {
				logrus.WithError(err).WithField("id", file.Id).Warn("failed to delete file")
			}
		}

		info.Committed = true
	}

	err = bundle.Close()
	if err != nil {
		return err
	}

	err = bundle.Compact()
	if err != nil {
		return err
	}

	if err := config.ModifyConfig(c.rhmConfigFlags.ConfigAccess(), *c.rhmRawConfig, true); err != nil {
		return err
	}

	return nil
}
