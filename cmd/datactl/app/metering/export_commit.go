// Copyright 2021 IBM Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metering

import (
	"context"
	"time"

	"github.com/gotidy/ptr"
	"github.com/redhat-marketplace/datactl/pkg/clients/dataservice"
	datactlapi "github.com/redhat-marketplace/datactl/pkg/datactl/api"
	dataservicev1 "github.com/redhat-marketplace/datactl/pkg/datactl/api/dataservice/v1"
	"github.com/redhat-marketplace/datactl/pkg/datactl/config"
	"github.com/redhat-marketplace/datactl/pkg/datactl/metering"
	"github.com/redhat-marketplace/datactl/pkg/datactl/output"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/kubectl/pkg/cmd/get"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	commitLong = templates.LongDesc(i18n.T(`
		Commits the file on the Dataservice.

		Committing indicates that the user will be delivering the files for
		processing by using the "{{ .cmd }} export push" command. Commiting files
		are recorded in the datactl config file.`))

	commitExample = templates.Examples(i18n.T(`
		# Commit all files in the active export file.
		{{ .cmd }} export commit

		# Run the commit but perform no actions (dry-run).
		{{ .cmd }} export commit --dry-run
`))
)

func NewCmdExportCommit(rhmFlags *config.ConfigFlags, f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := exportCommitOptions{
		rhmConfigFlags: rhmFlags,
		PrintFlags:     get.NewGetPrintFlags(),
		IOStreams:      ioStreams,
	}

	cmd := &cobra.Command{
		Use:                   "commit [(--dry-run)]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Finalizes the download of files."),
		Long:                  output.ReplaceCommandStrings(commitLong),
		Example:               output.ReplaceCommandStrings(commitExample),
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
	rhmConfigFlags *config.ConfigFlags
	PrintFlags     *get.PrintFlags

	dryRun bool

	//internal
	args        []string
	humanOutput bool

	rhmRawConfig *datactlapi.Config
	dataService  dataservice.Client

	currentMeteringExport *datactlapi.MeteringExport
	bundle                *metering.BundleFile

	ToPrinter func(string) (printers.ResourcePrinter, error)

	genericclioptions.IOStreams
}

func (c *exportCommitOptions) Complete(cmd *cobra.Command, args []string) error {
	c.args = args

	var err error
	c.rhmRawConfig, err = c.rhmConfigFlags.RawPersistentConfigLoader().RawConfig()
	if err != nil {
		return err
	}

	c.dataService, err = c.rhmConfigFlags.DataServiceClient()
	if err != nil {
		return err
	}

	c.currentMeteringExport, err = c.rhmConfigFlags.MeteringExport()
	if err != nil {
		return err
	}

	c.bundle, err = metering.NewBundleFromExport(c.currentMeteringExport)
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
	} else {
		output.DisableColor()
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

	p := output.NewHumanOutput()

	if c.dryRun {
		p = p.WithDetails("dryRun", true)
	}

	if c.humanOutput {
		p = p.WithDetails("cluster", c.currentMeteringExport.DataServiceCluster)
		p.Titlef("%s", i18n.T("commit started"))

		if c.dryRun {
			p.Warnf(i18n.T("dry-run enabled; files will not be committed"))
		}

		p = p.Sub()
		p.WithDetails("exportFile", c.currentMeteringExport.FileName).Infof(i18n.T("file commit status:"))
	}

	errs := map[string]error{}
	committed := 0

	for _, file := range c.currentMeteringExport.Files {
		file.Action = dataservicev1.Commit
		file.Result = dataservicev1.Ok

		if c.dryRun || file.Committed == true {
			if c.dryRun {
				file.Result = dataservicev1.DryRun
			}
			print.PrintObj(file, writer)
			writer.Flush()
			continue
		}

		err := c.dataService.DeleteFile(ctx, file.Id)
		if err != nil {
			file.Error = err.Error()
			file.Committed = false
			file.Result = dataservicev1.Error
			print.PrintObj(file, writer)
			writer.Flush()
			errs[file.Name] = err
			continue
		}

		file.Error = ""
		file.Committed = true
		committed = committed + 1
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

	if c.humanOutput {
		p.WithDetails("committed", committed, "files", len(c.currentMeteringExport.Files)).Infof(i18n.T("commit finished"))

		if len(errs) != 0 {
			p.Errorf(nil, "errors have occurred")
			p2 := p.Sub()
			for name, err := range errs {
				p2.WithDetails("name", name).Errorf(nil, err.Error())
			}
		}
	}

	// if dryRun, stop early
	if c.dryRun {
		return nil
	}

	if err := config.ModifyConfig(c.rhmConfigFlags.ConfigAccess(), *c.rhmRawConfig, true); err != nil {
		return err
	}

	return nil
}
