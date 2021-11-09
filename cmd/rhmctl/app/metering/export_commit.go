package metering

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/redhat-marketplace/rhmctl/pkg/clients/dataservice"
	rhmctlapi "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	"github.com/redhat-marketplace/rhmctl/pkg/rhmctl/config"
	"github.com/redhat-marketplace/rhmctl/pkg/rhmctl/metering"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
)

func NewCmdExportCommit(rhmFlags *config.ConfigFlags, f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := exportCommitOptions{
		configFlags:    genericclioptions.NewConfigFlags(false),
		rhmConfigFlags: rhmFlags,
	}

	cmd := &cobra.Command{
		Use:                   "commit",
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

	return cmd
}

type exportCommitOptions struct {
	configFlags    *genericclioptions.ConfigFlags
	rhmConfigFlags *config.ConfigFlags

	//internal
	args      []string
	rawConfig clientapi.Config

	rhmRawConfig *rhmctlapi.Config
	dataService  dataservice.Client

	currentMeteringExport *rhmctlapi.MeteringExport
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

	for _, export := range c.rhmRawConfig.MeteringExports {
		if export.Active == true {
			c.currentMeteringExport = export
		}
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

	for _, info := range c.currentMeteringExport.FileInfo {
		if info.Committed {
			continue
		}

		for _, f := range info.Files {
			err := c.dataService.DeleteFile(ctx, f.Id)
			if err != nil {
				logrus.WithError(err).WithField("id", f.Id).Warn("failed to delete file")
			}
		}

		info.Committed = true

		data, err := json.Marshal(info)
		w, err := bundle.NewFile("commit.json", int64(len(data)))
		if err != nil {
			return err
		}
		io.Copy(w, bytes.NewReader(data))
	}

	err = bundle.Close()
	if err != nil {
		return err
	}

	if err := config.ModifyConfig(c.rhmConfigFlags.ConfigAccess(), *c.rhmRawConfig, true); err != nil {
		return err
	}

	return nil
}
