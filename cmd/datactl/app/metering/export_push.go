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
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"emperror.dev/errors"
	"github.com/gotidy/ptr"
	"github.com/redhat-marketplace/datactl/pkg/bundle"
	"github.com/redhat-marketplace/datactl/pkg/clients/marketplace"
	datactlapi "github.com/redhat-marketplace/datactl/pkg/datactl/api"
	dataservicev1 "github.com/redhat-marketplace/datactl/pkg/datactl/api/dataservice/v1"
	"github.com/redhat-marketplace/datactl/pkg/datactl/config"
	"github.com/redhat-marketplace/datactl/pkg/printers/output"
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
		Pushes files to the metrics processing backends.

		Pushing uses the current kubernetes context and records the results into
		the datactl config file.`))

	pushExamples = templates.Examples(i18n.T(`
		# Push the files in the active export
	 	{{ .cmd }} export push

		# Run the push but perform no actions (dry-run).
		{{ .cmd }} export push --dry-run

		# Push a specific {{ .cmd }} file
		{{ .cmd }} export push --file={{ .defaultDataPath }}/rhm-upload-20211111T000959Z.tar
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
		Long:                  output.ReplaceCommandStrings(pushLong),
		Example:               output.ReplaceCommandStrings(pushExamples),
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

	rhmRawConfig *datactlapi.Config
	marketplace  marketplace.Client

	currentMeteringExport *datactlapi.MeteringExport
	bundle                *bundle.BundleFile

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

	e.marketplace, err = e.rhmConfigFlags.MarketplaceClient()
	if err != nil {
		return err
	}

	e.currentMeteringExport, err = e.rhmConfigFlags.MeteringExport()
	if err != nil {
		return err
	}

	e.bundle, err = bundle.NewBundleFromExport(e.currentMeteringExport)
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
	} else {
		output.DisableColor()
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

func (e *exportPushOptions) Run() error {
	// TODO make timeout configurable
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	writer := printers.GetNewTabWriter(e.Out)
	p := output.NewHumanOutput()

	print, err := e.ToPrinter("pushed")
	if err != nil {
		return err
	}

	print = output.NewActionCLITableOrStruct(e.Out, e.PrintFlags, print)

	file := e.currentMeteringExport.FileName

	if e.OverrideFile != "" {
		file = e.OverrideFile
		print = output.NewPushFileOnlyCLITableOrStruct(e.PrintFlags, print)
	}

	if e.humanOutput {
		p.WithDetails("uploadHost", e.rhmRawConfig.MarketplaceEndpoint.Host).
			Titlef(i18n.T("push started"))
		p = p.Sub()

		if e.dryRun {
			p.Warnf(i18n.T("dry-run enabled; files will not be pushed"))
		}

		p.WithDetails("exportFile", file).Infof(i18n.T("pushing files status:"))
	}

	files := map[string]*dataservicev1.FileInfoCTLAction{}

	if e.OverrideFile == "" {
		for _, f := range e.currentMeteringExport.Files {
			localF := f
			files[f.Name] = localF
		}
	}

	errs := map[string]error{}
	found := 0
	pushed := 0

	err = bundle.WalkTar(e.currentMeteringExport.FileName, func(header *tar.Header, r io.Reader) error {
		// skip our helper commit file
		if header.Name == "commit.json" {
			return nil
		}

		log := logger.WithValues("file", header.Name).V(5)

		file := files[header.Name]

		if file == nil && e.OverrideFile == "" {
			log.Info("tar file has no info in the config file skipping", "file", header.Name)
			return nil
		}

		if e.OverrideFile != "" {
			// handle the case where we are just uploading from a file and have no config
			file = &dataservicev1.FileInfoCTLAction{
				FileInfo: &dataservicev1.FileInfo{},
			}
			file.Name = header.Name
			file.Size = uint32(header.Size)
			file.Action = "Pushed"
		}

		found = found + 1

		file.Action = dataservicev1.Push
		file.Result = dataservicev1.Ok

		if e.dryRun || file.Pushed {
			if e.dryRun {
				file.Result = dataservicev1.DryRun
			}

			if file.Pushed {
				log.Info("file is already pushed")
			}

			print.PrintObj(file, writer)
			writer.Flush()
			return nil
		}

		id, err := e.marketplace.Metrics().Upload(ctx, header.Name, r)
		if err != nil {
			details := errors.GetDetails(err)
			err = errors.Errorf("%s %+v", err.Error(), details)
			log.Info("failed to push file", "err", err)
			errs[file.Name] = err
			file.Error = err.Error()
			file.Action = dataservicev1.Pull
			file.Result = dataservicev1.Error
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
		pushed = pushed + 1
		log.Info("push file success")
		return nil
	})

	if err != nil {
		return err
	}

	if e.humanOutput {
		p.WithDetails("pushed", pushed, "files", found).Infof(i18n.T("push finished"))

		if len(errs) != 0 {
			p.Errorf(nil, "errors have occurred")
			p2 := p.Sub()
			for name, err := range errs {
				p2.WithDetails("name", name).Errorf(nil, err.Error())
			}
		}
	}

	// if on dryrun, stop before we save
	if e.dryRun {
		return nil
	}

	// if we're reading from an override file, don't save or compact the file
	if e.OverrideFile != "" {
		return nil
	}

	err = e.bundle.Compact(nil)
	if err != nil {
		return err
	}

	if err := config.ModifyConfig(e.rhmConfigFlags.RawPersistentConfigLoader().ConfigAccess(), *e.rhmRawConfig, true); err != nil {
		return err
	}
	return nil
}
