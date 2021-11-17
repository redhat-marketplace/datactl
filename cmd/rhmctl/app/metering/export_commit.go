package metering

import (
	"context"
	"fmt"
	"time"

	"github.com/gotidy/ptr"
	"github.com/redhat-marketplace/rhmctl/pkg/clients/dataservice"
	rhmctlapi "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	"github.com/redhat-marketplace/rhmctl/pkg/rhmctl/config"
	"github.com/redhat-marketplace/rhmctl/pkg/rhmctl/metering"
	"github.com/redhat-marketplace/rhmctl/pkg/rhmctl/output"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/kubectl/pkg/cmd/get"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	commitLong = templates.LongDesc(i18n.T(`
		Commits the file on the Red Hat Marketplace Dataservice.

		Committing indicates that the user will be delivering the files for
		processing by using the "%[1]s export push" command. Commiting files
		are recorded in the rhmctl config file.`))

	commitExample = templates.Examples(i18n.T(`
		# Commit all files in the active export file.
		%[1]s export commit

		# Run the commit but perform no actions (dry-run).
		%[1]s export commit --dry-run
`))
)

func NewCmdExportCommit(rhmFlags *config.ConfigFlags, f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	pathOptions := genericclioptions.NewConfigFlags(false)
	o := exportCommitOptions{
		configFlags:    pathOptions,
		rhmConfigFlags: rhmFlags,
		PrintFlags:     get.NewGetPrintFlags(),
		IOStreams:      ioStreams,
	}

	cmd := &cobra.Command{
		Use:                   "commit [(--dry-run)]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Finalizes the download of files."),
		Long:                  fmt.Sprintf(commitLong, output.CommandName()),
		Example:               fmt.Sprintf(commitExample, output.CommandName()),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}
	o.PrintFlags.AddFlags(cmd)
	cmd.Flags().MarkHidden("label-columns")
	cmd.Flags().MarkHidden("sort-by")
	cmd.Flags().MarkHidden("show-kind")
	cmd.Flags().MarkHidden("show-managed-fields")
	cmd.Flags().MarkHidden("show-labels")

	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, i18n.T("No action taken. Print only."))
	return cmd
}

type exportCommitOptions struct {
	configFlags    *genericclioptions.ConfigFlags
	rhmConfigFlags *config.ConfigFlags
	PrintFlags     *get.PrintFlags

	dryRun bool

	//internal
	args        []string
	humanOutput bool
	rawConfig   clientapi.Config

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

	c.currentMeteringExport, c.bundle, err = createOrUpdateBundle(c.rawConfig.CurrentContext, c.rhmRawConfig)
	if err != nil {
		return err
	}

	c.ToPrinter = func(operation string) (printers.ResourcePrinter, error) {
		c.PrintFlags.NamePrintFlags.Operation = operation
		return c.PrintFlags.ToPrinter()
	}

	if c.PrintFlags.OutputFormat == nil || *c.PrintFlags.OutputFormat == "wide" || *c.PrintFlags.OutputFormat == "" {
		c.humanOutput = true
		c.PrintFlags.OutputFormat = ptr.String("wide")
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

	writer := printers.GetNewTabWriter(c.Out)

	print = output.NewActionCLITableOrStruct(c.PrintFlags, print)

	if c.dryRun {
		logrus.Info(i18n.T("dry-run enabled, files will not be removed from data service"))
	}

	for _, file := range c.currentMeteringExport.Files {
		if c.dryRun || file.Committed == true {
			file.Action = "Commit"
			print.PrintObj(file, writer)
			writer.Flush()
			continue
		}

		err := c.dataService.DeleteFile(ctx, file.Id)
		if err != nil {
			logrus.WithError(err).WithField("id", file.Id).Warn("failed to delete file")
			file.Error = err.Error()
			file.Committed = false
			file.Action = "Error"
			print.PrintObj(file, writer)
			writer.Flush()
			continue
		}

		file.Error = ""
		file.Action = "Commit"
		file.Committed = true
		print.PrintObj(file, writer)
		writer.Flush()
	}

	err = bundle.Close()
	if err != nil {
		return err
	}

	err = bundle.Compact(nil)
	if err != nil {
		return err
	}

	writer.Flush()

	// if dryRun, stop early
	if c.dryRun {
		return nil
	}

	if err := config.ModifyConfig(c.rhmConfigFlags.ConfigAccess(), *c.rhmRawConfig, true); err != nil {
		return err
	}

	return nil
}
