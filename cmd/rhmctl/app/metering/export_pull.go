package metering

import (
	"context"
	"fmt"
	"time"

	"emperror.dev/errors"
	"github.com/gotidy/ptr"
	"github.com/redhat-marketplace/rhmctl/pkg/clients/dataservice"
	rhmctlapi "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	"github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api/latest"
	"github.com/redhat-marketplace/rhmctl/pkg/rhmctl/config"
	"github.com/redhat-marketplace/rhmctl/pkg/rhmctl/metering"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
)

func NewCmdExportPull(rhmFlags *config.ConfigFlags, f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	pathOptions := genericclioptions.NewConfigFlags(false)
	o := exportPullOptions{
		configFlags:    pathOptions,
		rhmConfigFlags: rhmFlags,
		PrintFlags:     genericclioptions.NewPrintFlags("export pull").WithTypeSetter(latest.Scheme),
		IOStreams:      ioStreams,
	}

	cmd := &cobra.Command{
		Use:                   "pull [(--before DATE) (--after DATE) (--include-deleted)]",
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

	o.PrintFlags.AddFlags(cmd)
	cmdutil.AddDryRunFlag(cmd)
	cmd.Flags().BoolVar(&o.includeDeleted, "include-deleted", false, "include deleted files")

	return cmd
}

type exportPullOptions struct {
	configFlags    *genericclioptions.ConfigFlags
	rhmConfigFlags *config.ConfigFlags
	PrintFlags     *genericclioptions.PrintFlags

	//flags
	includeDeleted bool

	//internal
	args      []string
	rawConfig clientapi.Config

	rhmRawConfig *rhmctlapi.Config
	dataService  dataservice.Client

	ToPrinter func(string) (printers.ResourcePrinter, error)

	bundle                *metering.BundleFile
	currentMeteringExport *rhmctlapi.MeteringExport

	genericclioptions.IOStreams
}

func (e *exportPullOptions) Complete(cmd *cobra.Command, args []string) error {
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

	//cmdutil.PrintFlagsWithDryRunStrategy(o.PrintFlags, o.dryRunStrategy)
	e.ToPrinter = func(operation string) (printers.ResourcePrinter, error) {
		e.PrintFlags.NamePrintFlags.Operation = operation

		return e.PrintFlags.ToPrinter()
	}

	for _, export := range e.rhmRawConfig.MeteringExports {
		fmt.Println(export)
		if export.Active {
			localExport := *export
			e.currentMeteringExport = &localExport
		}
	}

	if e.currentMeteringExport == nil {
		bundle, err := metering.NewBundleWithDefaultName()
		if err != nil {
			return errors.Wrap(err, "creating bundle")
		}

		e.currentMeteringExport = &rhmctlapi.MeteringExport{
			FileName: bundle.Name(),
			Active:   true,
			Start:    metav1.Timestamp{Seconds: metav1.Now().Unix()},
		}

		e.rhmRawConfig.MeteringExports[bundle.Name()] = e.currentMeteringExport
		e.bundle = bundle
	}

	return nil
}

func (e *exportPullOptions) Validate() error {
	if e.currentMeteringExport == nil || e.currentMeteringExport.FileName == "" {
		return errors.New("command requires a current export file")
	}

	if e.bundle == nil {
		return errors.New("command requires a current export file")
	}

	return nil
}

func (e *exportPullOptions) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	defer e.bundle.Close()

	response := rhmctlapi.ListFilesResponse{}
	listOpts := dataservice.ListOptions{
		IncludeDeleted: e.includeDeleted,
	}

	exportInfo := rhmctlapi.MeteringFileSummary{}
	exportInfo.DataServiceContext = e.rawConfig.CurrentContext
	exportInfo.Committed = false
	exportInfo.Files = make([]*rhmctlapi.FileInfo, 0)

	// print, err := e.ToPrinter("download")
	// if err != nil {
	// 	return err
	// }

	for {
		err := e.dataService.ListFiles(ctx, listOpts, &response)

		if err != nil {
			return err
		}

		fmt.Printf("%+v", response.Files)

		for _, file := range response.Files {
			w, err := e.bundle.NewFile(file.Name, int64(file.Size))
			if err != nil {
				return err
			}

			exportInfo.Files = append(exportInfo.Files, file)

			_, err = e.dataService.DownloadFile(ctx, file.Id, w)
			if err != nil {
				return err
			}

			//print.PrintObj(file, e.Out)
		}

		if response.NextPageToken == "" {
			break
		}

		listOpts.PageSize = ptr.Int(int(response.PageSize))
		listOpts.PageToken = response.NextPageToken
	}

	err := e.bundle.Close()
	if err != nil {
		return err
	}

	fmt.Println(e.bundle.Name())

	if err := config.ModifyConfig(e.rhmConfigFlags.RawPersistentConfigLoader().ConfigAccess(), *e.rhmRawConfig, true); err != nil {
		return err
	}
	// TODO: save the metering export data

	return nil
}
