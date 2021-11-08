package metering

import (
	"context"
	"time"

	"emperror.dev/errors"
	"github.com/gotidy/ptr"
	"github.com/redhat-marketplace/rhmctl/pkg/clients"
	"github.com/redhat-marketplace/rhmctl/pkg/clients/dataservice"
	rhmctlapi "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	"github.com/redhat-marketplace/rhmctl/pkg/rhmctl/config"
	"github.com/redhat-marketplace/rhmctl/pkg/rhmctl/metering"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
)

func NewCmdExportPull(conf *rhmctlapi.Config, f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := exportPullOptions{
		configFlags: genericclioptions.NewConfigFlags(false),
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
			cmd.Help()
		},
	}

	cmd.Flags().BoolVar(&o.includeDeleted, "include-deleted", false, "include deleted files")
	o.configFlags.AddFlags(cmd.Flags())

	return cmd
}

type exportPullOptions struct {
	configFlags *genericclioptions.ConfigFlags

	//flags
	includeDeleted bool

	//internal
	args      []string
	rawConfig clientapi.Config

	rhmRawConfig *rhmctlapi.Config
	dataService  dataservice.Client
}

func (e *exportPullOptions) Complete(cmd *cobra.Command, args []string) error {
	e.args = args

	var err error
	e.rawConfig, err = e.configFlags.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return err
	}

	e.rhmRawConfig, err = config.LoadConfig(&config.DefaultLoadingRules{})
	if err != nil {
		return err
	}

	e.dataService, err = clients.ProvideDataService(e.rawConfig.CurrentContext, e.rhmRawConfig)

	if err != nil {
		return err
	}

	return nil
}

func (e *exportPullOptions) Validate() error {
	if e.rhmRawConfig.CurrentMeteringExport == nil {
		return errors.New("command requires a current export; run `rhmctl export start`")
	}

	if e.rhmRawConfig.CurrentMeteringExport.FileName == "" {
		return errors.New("command requires a current export file")
	}

	return nil
}

func (e *exportPullOptions) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	bundle, err := metering.NewBundle(e.rhmRawConfig.CurrentMeteringExport.FileName)
	if err != nil {
		return err
	}

	response := rhmctlapi.ListFilesResponse{}
	listOpts := dataservice.ListOptions{
		IncludeDeleted: e.includeDeleted,
	}

	exportInfo := &rhmctlapi.MeteringFileSummary{}
	exportInfo.DataServiceContext = e.rawConfig.CurrentContext
	exportInfo.Committed = false
	exportInfo.Files = make([]*rhmctlapi.FileInfo, 0)

	for {
		err := e.dataService.ListFiles(ctx, listOpts, &response)

		if err != nil {
			return err
		}

		if response.NextPageToken == "" {
			break
		}

		for _, file := range response.Files {
			w, err := bundle.NewFile(file.Name, int64(file.Size))
			if err != nil {
				return err
			}

			exportInfo.Files = append(exportInfo.Files, file)

			_, err = e.dataService.DownloadFile(ctx, file.Id, w)
			if err != nil {
				return err
			}
		}

		listOpts.PageSize = ptr.Int(int(response.PageSize))
		listOpts.PageToken = response.NextPageToken
	}

	// TODO: save the config file
	// TODO: save the metering export data

	return nil
}
