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

	"emperror.dev/errors"
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
	clientapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/kubectl/pkg/cmd/get"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	pullLong = templates.LongDesc(i18n.T(`
		Pulls files from the Dataservice on the cluster.

		Prints a table of the files pulled with basic information. The --before or --after flags
		can be used to change the date range that the files are pulled from. All dates must be in
		RFC3339 format as defined by the Golang time package.

		If the files have already been pulled then using the --include-deleted flag may be necessary.`))

	pullExample = templates.Examples(i18n.T(`
		# Pull all available files from the current dataservice cluster to Usage
		{{ .cmd }} export pull

		# Pull all files before November 14th, 2021
		{{ .cmd }} export pull --before 2021-11-15T00:00:00Z

		# Pull all files after November 14th, 2021
		{{ .cmd }} export pull

		# Pull all files between November 14th, 2021 and November 15th, 2021
		{{ .cmd }} export pull --after 2021-11-14T00:00:00Z --before 2021-11-15T00:00:00Z

		# Pull all deleted files
		{{ .cmd }} export pull --include-deleted`))
)

func NewCmdExportPull(rhmFlags *config.ConfigFlags, f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := exportPullOptions{
		rhmConfigFlags: rhmFlags,
		PrintFlags:     get.NewGetPrintFlags(),
		IOStreams:      ioStreams,
	}

	cmd := &cobra.Command{
		Use:                   "pull [(--before DATE) (--after DATE) (--include-deleted)]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Pulls files from RHM Operator"),
		Long:                  output.ReplaceCommandStrings(pullLong),
		Example:               output.ReplaceCommandStrings(pullExample),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}

	o.PrintFlags.AddFlags(cmd)
	cmd.Flags().BoolVar(&o.includeDeleted, "include-deleted", false, i18n.T("include deleted files"))
	cmd.Flags().StringVar(&o.beforeDate, "before", "", i18n.T("pull files before date"))
	cmd.Flags().StringVar(&o.afterDate, "after", "", i18n.T("pull files after date"))

	cmd.Flags().MarkHidden("label-columns")
	cmd.Flags().MarkHidden("sort-by")
	cmd.Flags().MarkHidden("show-kind")
	cmd.Flags().MarkHidden("show-managed-fields")
	cmd.Flags().MarkHidden("show-labels")

	return cmd
}

type exportPullOptions struct {
	rhmConfigFlags *config.ConfigFlags
	PrintFlags     *get.PrintFlags

	//flags
	includeDeleted        bool
	beforeDate, afterDate string

	//derivedFlags
	humanOutput             bool
	beforeDateT, afterDateT time.Time

	//internal
	args      []string
	rawConfig clientapi.Config

	rhmRawConfig *datactlapi.Config
	dataService  dataservice.Client

	ToPrinter func(string) (printers.ResourcePrinter, error)

	bundle                *metering.BundleFile
	currentMeteringExport *datactlapi.MeteringExport
	clusterName           string

	genericclioptions.IOStreams
}

func (e *exportPullOptions) Complete(cmd *cobra.Command, args []string) error {
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

	e.ToPrinter = func(operation string) (printers.ResourcePrinter, error) {
		e.PrintFlags.NamePrintFlags.Operation = operation
		return e.PrintFlags.ToPrinter()
	}

	e.currentMeteringExport, err = e.rhmConfigFlags.MeteringExport()
	if err != nil {
		return err
	}

	e.bundle, err = metering.NewBundleFromExport(e.currentMeteringExport)
	if err != nil {
		return err
	}

	if e.PrintFlags.OutputFormat == nil || *e.PrintFlags.OutputFormat == "wide" || *e.PrintFlags.OutputFormat == "" {
		e.humanOutput = true
		output.SetOutput(e.Out, true)
		e.PrintFlags.OutputFormat = ptr.String("wide")
	}

	if e.beforeDate != "" {
		e.beforeDateT, err = time.Parse(time.RFC3339, e.beforeDate)
		if err != nil {
			return errors.Wrapf(err, "provided before time %s does not fit into RFC3339 layout %s", e.beforeDate, time.RFC3339)
		}
	}

	if e.afterDate != "" {
		e.afterDateT, err = time.Parse(time.RFC3339, e.afterDate)
		if err != nil {
			return errors.Wrapf(err, "provided after time %s does not fit into RFC3339 layout %s", e.afterDate, time.RFC3339)
		}
	}

	return nil
}

func (e *exportPullOptions) Validate() error {
	if e.currentMeteringExport == nil || e.currentMeteringExport.FileName == "" {
		return errors.New("command requires a current export file")
	}

	if e.bundle == nil {
		return errors.New("command requires a current export bundle file")
	}

	return nil
}

func (e *exportPullOptions) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	defer e.bundle.Close()

	response := dataservicev1.ListFilesResponse{}
	listOpts := dataservice.ListOptions{
		IncludeDeleted: e.includeDeleted,
		BeforeDate:     e.beforeDateT,
		AfterDate:      e.afterDateT,
	}

	if e.currentMeteringExport.Files == nil {
		e.currentMeteringExport.Files = make([]*dataservicev1.FileInfoCTLAction, 0)
	}

	p := output.NewHumanOutput()

	if e.humanOutput {
		p.WithDetails("cluster", e.currentMeteringExport.DataServiceCluster).
			Titlef("%s", i18n.T("pull started"))
		p = p.Sub()
		p.WithDetails("exportFile", e.currentMeteringExport.FileName).Infof(i18n.T("files pulled status:"))
	}

	writer := printers.GetNewTabWriter(e.Out)

	print, err := e.ToPrinter("pulled")
	if err != nil {
		return err
	}

	print = output.NewActionCLITableOrStruct(e.PrintFlags, print)

	files := []*dataservicev1.FileInfoCTLAction{}
	errs := map[string]error{}
	found := 0
	pulled := 0

	for {
		err := e.dataService.ListFiles(ctx, listOpts, &response)

		if err != nil {
			return err
		}

		for i := range response.Files {
			cliFile := dataservicev1.NewFileInfoCTLAction(response.Files[i])
			files = append(files, cliFile)
			found = found + 1

			w, err := e.bundle.NewFile(cliFile.Name, int64(cliFile.Size))
			if err != nil {
				return err
			}

			_, err = e.dataService.DownloadFile(ctx, cliFile.Id, w)
			if err != nil {
				cliFile.Action = dataservicev1.Pull
				cliFile.Result = dataservicev1.Error
				cliFile.Error = err.Error()
				errs[cliFile.Name] = err

				print.PrintObj(cliFile, writer)
				writer.Flush()
				continue
			}

			cliFile.Action = dataservicev1.Pull
			cliFile.Result = dataservicev1.Ok
			pulled = pulled + 1
			print.PrintObj(cliFile, writer)
			writer.Flush()
		}

		if response.NextPageToken == "" {
			break
		}

		listOpts.PageSize = ptr.Int(int(response.PageSize))
		listOpts.PageToken = response.NextPageToken
	}

	filesMap := map[string]*dataservicev1.FileInfoCTLAction{}
	fileNames := map[string]interface{}{}

	for _, f := range e.currentMeteringExport.Files {
		if f.Committed && f.Pushed {
			continue
		}

		filesMap[f.Name+f.Source+f.SourceType] = f
		fileNames[f.Name] = nil
	}

	for _, f := range files {
		filesMap[f.Name+f.Source+f.SourceType] = f
		fileNames[f.Name] = nil
	}

	e.currentMeteringExport.Files = []*dataservicev1.FileInfoCTLAction{}

	for i := range filesMap {
		e.currentMeteringExport.Files = append(e.currentMeteringExport.Files, filesMap[i])
	}

	err = e.bundle.Close()
	if err != nil {
		return err
	}

	err = e.bundle.Compact(fileNames)
	if err != nil {
		return err
	}

	if err := config.ModifyConfig(e.rhmConfigFlags.RawPersistentConfigLoader().ConfigAccess(), *e.rhmRawConfig, true); err != nil {
		return err
	}

	if found == 0 {
		return i18n.Errorf("no files found")
	}

	if e.humanOutput {
		p.WithDetails("found", found, "pulled", pulled).Infof(i18n.T("pull complete"))

		if len(errs) != 0 {
			p.Errorf(nil, "errors have occurred")
			p2 := p.Sub()
			for name, err := range errs {
				p2.WithDetails("name", name).Errorf(nil, err.Error())
			}
		}
	}

	return nil
}
