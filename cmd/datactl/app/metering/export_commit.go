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

	"github.com/redhat-marketplace/datactl/pkg/bundle"
	"github.com/redhat-marketplace/datactl/pkg/clients/dataservice"
	datactlapi "github.com/redhat-marketplace/datactl/pkg/datactl/api"
	"github.com/redhat-marketplace/datactl/pkg/datactl/config"
	"github.com/redhat-marketplace/datactl/pkg/printers"
	"github.com/redhat-marketplace/datactl/pkg/printers/output"
	"github.com/redhat-marketplace/datactl/pkg/sources"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
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
	bundle                *bundle.BundleFile

	printer printers.Printer

	genericclioptions.IOStreams

	sources.Factory
}

func (c *exportCommitOptions) Complete(cmd *cobra.Command, args []string) error {
	c.args = args

	var err error
	c.rhmRawConfig, err = c.rhmConfigFlags.RawPersistentConfigLoader().RawConfig()
	if err != nil {
		return err
	}

	c.currentMeteringExport, err = c.rhmConfigFlags.MeteringExport()
	if err != nil {
		return err
	}

	c.bundle, err = bundle.NewBundleFromExport(c.currentMeteringExport)
	if err != nil {
		return err
	}

	c.PrintFlags.NamePrintFlags.Operation = "commit"

	c.printer, err = printers.NewPrinter(c.Out, c.PrintFlags)

	if err != nil {
		return err
	}

	c.Factory = (&sources.SourceFactoryBuilder{}).
		SetConfigFlags(c.rhmConfigFlags).
		SetPrinter(c.printer).
		Build()

	return nil
}

func (c *exportCommitOptions) Validate() error {
	return nil
}

func (c *exportCommitOptions) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	defer c.bundle.Close()

	c.printer.HumanOutput(func(ho *output.HumanOutput) *output.HumanOutput {
		p := ho

		if c.dryRun {
			p = ho.WithDetails("dryRun", true)
		}

		p = p.WithDetails("cluster", c.currentMeteringExport.DataServiceCluster)
		p.Titlef("%s", i18n.T("commit started"))

		if c.dryRun {
			p.Warnf(i18n.T("dry-run enabled; files will not be committed"))
		}

		p = p.Sub()
		p.WithDetails("exportFile", c.currentMeteringExport.FileName).Infof(i18n.T("file commit status:"))
		return p
	})

	errs := []error{}
	committed := 0

	for name := range c.rhmRawConfig.Sources {
		s := c.rhmRawConfig.Sources[name]
		source, err := c.Factory.FromSource(*s)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		commitSource, ok := source.(sources.CommitableSource)

		if !ok {
			continue
		}

		c.printer.HumanOutput(func(p *output.HumanOutput) *output.HumanOutput {
			p = p.WithDetails("sourceName", s.Name, "sourceType", s.Type)
			p.Titlef("%s", i18n.T("commit started for source"))
			return p
		})

		count, err := commitSource.Commit(ctx, c.currentMeteringExport, c.bundle, sources.EmptyOptions())
		if err != nil {
			errs = append(errs, err)
		}
		committed += count

		c.printer.HumanOutput(func(p *output.HumanOutput) *output.HumanOutput {
			p.WithDetails("count", count).Infof(i18n.T("commit complete"))
			return p
		})
	}

	err := c.bundle.Compact(nil)
	if err != nil {
		return err
	}

	c.printer.HumanOutput(func(ho *output.HumanOutput) *output.HumanOutput {
		p := ho
		p.WithDetails("committed", committed, "files", len(c.currentMeteringExport.Files)).Infof(i18n.T("commit finished"))

		if len(errs) != 0 {
			p.Errorf(nil, "errors have occurred")
		}
		return p
	})

	// if dryRun, stop early
	if c.dryRun {
		return nil
	}

	if err := config.ModifyConfig(c.rhmConfigFlags.ConfigAccess(), *c.rhmRawConfig, true); err != nil {
		return err
	}

	return nil
}
