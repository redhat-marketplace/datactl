package metering

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"emperror.dev/errors"
	"github.com/gotidy/ptr"
	"github.com/redhat-marketplace/rhmctl/pkg/clients/dataservice"
	"github.com/redhat-marketplace/rhmctl/pkg/clients/marketplace"
	rhmctlapi "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	dataservicev1 "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api/dataservice/v1"
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
	pushLong = templates.LongDesc(i18n.T(`
		Pushes files to the Red Hat Marketplace metrics processing backends.

		Pushing uses the current kubernetes context and records the results into
		the rhmctl config file.`))

	pushExamples = templates.Examples(i18n.T(`
		# Push the files in the active export
	 	%[1]s export pull

		# Run the push but perform no actions (dry-run).
		%[1]s export pull --dry-run

		# Push a specific rhmctl file
		%[1]s export pull --file=$HOME/.rhmctl/data/rhm-upload-20211111T000959Z.tar
`))
)

func NewCmdExportPush(rhmFlags *config.ConfigFlags, f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := exportPushOptions{
		rhmConfigFlags: rhmFlags,
		PrintFlags:     get.NewGetPrintFlags(),
		IOStreams:      ioStreams,
	}

	cmd := &cobra.Command{
		Use:                   "push [(--dry-run)]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Pushes commited files."),
		Long:                  fmt.Sprintf(pushLong, output.CommandName()),
		Example:               fmt.Sprintf(pushExamples, output.CommandName()),
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

	cmd.Flags().StringVar(&o.OverrideFile, "file", "", i18n.T("tar file to upload from"))
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, i18n.T("No action taken. Print only."))

	return cmd
}

type exportPushOptions struct {
	rhmConfigFlags *config.ConfigFlags
	PrintFlags     *get.PrintFlags

	// Flags
	dryRun       bool
	OverrideFile string

	//internal
	humanOutput bool
	args        []string
	rawConfig   clientapi.Config

	rhmRawConfig *rhmctlapi.Config
	dataService  dataservice.Client
	marketplace  marketplace.Client

	currentMeteringExport *rhmctlapi.MeteringExport
	bundle                *metering.BundleFile

	ToPrinter func(string) (printers.ResourcePrinter, error)

	genericclioptions.IOStreams
}

func (e *exportPushOptions) Complete(cmd *cobra.Command, args []string) error {
	e.args = args

	var err error
	e.rhmRawConfig, err = e.rhmConfigFlags.RawPersistentConfigLoader().RawConfig()
	if err != nil {
		return err
	}

	e.dataService, err = e.rhmConfigFlags.DataServiceClient()
	if err != nil {
		return err
	}

	e.marketplace, err = e.rhmConfigFlags.MarketplaceClient()
	if err != nil {
		return err
	}

	e.currentMeteringExport, err = e.rhmConfigFlags.MeteringExport()
	if err != nil {
		return err
	}

	e.bundle, err = metering.NewBundleFromExport(e.currentMeteringExport)
	if err != nil {
		return err
	}

	e.ToPrinter = func(operation string) (printers.ResourcePrinter, error) {
		e.PrintFlags.NamePrintFlags.Operation = operation
		return e.PrintFlags.ToPrinter()
	}

	if e.PrintFlags.OutputFormat == nil || *e.PrintFlags.OutputFormat == "wide" || *e.PrintFlags.OutputFormat == "" {
		e.humanOutput = true
		e.PrintFlags.OutputFormat = ptr.String("wide")
	}

	return nil
}

func (e *exportPushOptions) Validate() error {
	if e.OverrideFile != "" {
		if _, err := os.Stat(e.OverrideFile); os.IsNotExist(err) {
			return fmt.Errorf("file does not exist %s", e.OverrideFile)
		}
	}

	return nil
}

func (e *exportPushOptions) runFileOnly(ctx context.Context) error {
	if e.dryRun {
		logrus.Warn(i18n.T("dry-run enabled, files will not be removed from data service"))
	}

	writer := printers.GetNewTabWriter(e.Out)

	print, err := e.ToPrinter("push")
	if err != nil {
		return err
	}

	print = output.NewPushFileOnlyCLITableOrStruct(e.PrintFlags, print)

	if e.dryRun {
		logrus.Warn(i18n.T("dry-run enabled, files will not be removed from data service"))
	}

	err = metering.WalkTar(e.OverrideFile, func(header *tar.Header, r io.Reader) error {
		// skip our helper commit file
		if header.Name == "commit.json" {
			return nil
		}

		file := &dataservicev1.FileInfoCTLAction{
			FileInfo: &dataservicev1.FileInfo{},
		}
		file.Name = header.Name
		file.Size = uint32(header.Size)
		file.Action = "Pushed"

		if e.dryRun {
			print.PrintObj(file, writer)
			writer.Flush()
			return nil
		}

		id, err := e.marketplace.Metrics().Upload(ctx, header.Name, r)
		if err != nil {
			details := errors.GetDetails(err)
			err = errors.Errorf("%s %+v", err.Error(), details)
			file.Error = err.Error()
			file.Pushed = false
			print.PrintObj(file, writer)
			writer.Flush()
			return nil
		}

		file.UploadError = ""
		file.Error = ""
		file.Pushed = true
		file.UploadID = id
		print.PrintObj(file, writer)
		writer.Flush()
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (e *exportPushOptions) Run() error {
	// TODO make timeout configurable
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if e.OverrideFile != "" {
		return e.runFileOnly(ctx)
	}

	bundle, err := metering.NewBundle(e.currentMeteringExport.FileName)
	if err != nil {
		return err
	}
	defer bundle.Close()

	writer := printers.GetNewTabWriter(e.Out)

	print, err := e.ToPrinter("pushed")
	if err != nil {
		return err
	}

	print = output.NewActionCLITableOrStruct(e.PrintFlags, print)

	if e.dryRun {
		logrus.Warn(i18n.T("dry-run enabled, files will not be removed from data service"))
	}

	files := map[string]*dataservicev1.FileInfoCTLAction{}
	for _, f := range e.currentMeteringExport.Files {
		localF := f
		files[f.Name] = localF
	}

	err = metering.WalkTar(e.currentMeteringExport.FileName, func(header *tar.Header, r io.Reader) error {
		// skip our helper commit file
		if header.Name == "commit.json" {
			return nil
		}

		file := files[header.Name]

		if file == nil {
			logrus.Warnf("tar file (%s) has no info in the config file skipping", header.Name)
			return nil
		}

		if file.Pushed {
			logrus.Debug("file has already been pushed")
			return nil
		}

		if e.dryRun {
			print.PrintObj(file, writer)
			writer.Flush()
			return nil
		}

		id, err := e.marketplace.Metrics().Upload(ctx, header.Name, r)
		if err != nil {
			details := errors.GetDetails(err)
			err = errors.Errorf("%s %+v", err.Error(), details)
			file.Error = err.Error()
			file.Pushed = false
			print.PrintObj(file, writer)
			writer.Flush()
			return nil
		}

		file.UploadError = ""
		file.Error = ""
		file.Pushed = true
		file.UploadID = id
		print.PrintObj(file, writer)
		writer.Flush()
		return nil
	})

	if err != nil {
		return err
	}

	// if on dryrun, stop before we save
	if e.dryRun {
		return nil
	}

	err = bundle.Compact(nil)
	if err != nil {
		return err
	}

	if err := config.ModifyConfig(e.rhmConfigFlags.RawPersistentConfigLoader().ConfigAccess(), *e.rhmRawConfig, true); err != nil {
		return err
	}

	return nil
}
