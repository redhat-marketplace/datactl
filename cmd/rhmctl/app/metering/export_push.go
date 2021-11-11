package metering

import (
	"archive/tar"
	"context"
	"io"
	"time"

	"emperror.dev/errors"
	"github.com/redhat-marketplace/rhmctl/pkg/clients/dataservice"
	"github.com/redhat-marketplace/rhmctl/pkg/clients/marketplace"
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

type exportPushOptions struct {
	configFlags    *genericclioptions.ConfigFlags
	rhmConfigFlags *config.ConfigFlags
	PrintFlags     *genericclioptions.PrintFlags

	dryRun bool

	//internal
	args      []string
	rawConfig clientapi.Config

	rhmRawConfig *rhmctlapi.Config
	dataService  dataservice.Client
	marketplace  marketplace.Client

	currentMeteringExport *rhmctlapi.MeteringExport
	bundle                *metering.BundleFile
	fileSummary           *rhmctlapi.MeteringFileSummary

	ToPrinter func(string) (printers.ResourcePrinter, error)

	genericclioptions.IOStreams
}

func NewCmdExportPush(rhmFlags *config.ConfigFlags, f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	pathOptions := genericclioptions.NewConfigFlags(false)
	o := exportPushOptions{
		configFlags:    pathOptions,
		rhmConfigFlags: rhmFlags,
		PrintFlags:     genericclioptions.NewPrintFlags("export pull").WithTypeSetter(latest.Scheme),
		IOStreams:      ioStreams,
	}

	cmd := &cobra.Command{
		Use:                   "push [(--dry-run)]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Pushes commited files."),
		//Long:                  imageLong,
		//Example:               imageExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}

	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, i18n.T("No action taken. Print only."))
	return cmd
}

func (e *exportPushOptions) Complete(cmd *cobra.Command, args []string) error {
	e.args = args

	var err error
	e.rawConfig, err = e.configFlags.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return err
	}

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

	e.currentMeteringExport, e.bundle, err = createOrUpdateBundle(e.rhmRawConfig)
	if err != nil {
		return err
	}

	e.ToPrinter = func(operation string) (printers.ResourcePrinter, error) {
		e.PrintFlags.NamePrintFlags.Operation = operation
		return e.PrintFlags.ToPrinter()
	}

	e.fileSummary = config.LatestContextMeteringFileSummary(e.currentMeteringExport, e.rawConfig.CurrentContext)

	return nil
}

func (e *exportPushOptions) Validate() error {
	return nil
}

func (e *exportPushOptions) Run() error {
	// TODO make timeout configurable
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	bundle, err := metering.NewBundle(e.currentMeteringExport.FileName)
	if err != nil {
		return err
	}
	defer bundle.Close()

	print, err := e.ToPrinter("pushed")
	if err != nil {
		return err
	}

	if e.dryRun {
		logrus.Warn(i18n.T("dry-run enabled, files will not be removed from data service"))
	}

	files := map[string]*rhmctlapi.FileInfo{}
	for _, info := range e.currentMeteringExport.FileInfo {
		for _, f := range info.Files {
			localF := f
			files[f.Name] = localF
		}
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

		if e.dryRun {
			print.PrintObj(file, e.Out)
			return nil
		}

		id, err := e.marketplace.Metrics().Upload(ctx, header.Name, r)
		if err != nil {
			details := errors.GetDetails(err)
			err = errors.Errorf("%s %+v", err.Error(), details)
			file.UploadError = err.Error()
			return nil
		}

		file.UploadID = id
		print.PrintObj(file, e.Out)
		return nil
	})

	if err != nil {
		return err
	}

	e.fileSummary.Pushed = true

	if err := config.ModifyConfig(e.rhmConfigFlags.RawPersistentConfigLoader().ConfigAccess(), *e.rhmRawConfig, true); err != nil {
		return err
	}

	return nil
}
