package metering

import (
	"context"
	"time"

	"github.com/redhat-marketplace/datactl/pkg/bundle"
	datactlapi "github.com/redhat-marketplace/datactl/pkg/datactl/api"
	"github.com/redhat-marketplace/datactl/pkg/datactl/config"
	"github.com/redhat-marketplace/datactl/pkg/printers"
	"github.com/redhat-marketplace/datactl/pkg/printers/output"
	"github.com/redhat-marketplace/datactl/pkg/sources"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/kubectl/pkg/cmd/get"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	pullLong = templates.LongDesc(i18n.T(`
		Pulls data from all available sources. Filtering by source name and type is available.

		Prints a table of the files pulled with basic information.

		Please use the sources commands to add new sources for pulling.`))

	pullExample = templates.Examples(i18n.T(`
		# Pull all available data from all available sources.
		{{ .cmd }} export pull all

		# Pull all data from a particular source-type.
		{{ .cmd }} export pull all --source-type dataService

		# Pull all data from a particular source.
		{{ .cmd }} export pull all --source-name my-dataservice-cluster
`))
)

func NewCmdExportPull(rhmFlags *config.ConfigFlags, f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := exportPullOptions{
		rhmConfigFlags: rhmFlags,
		PrintFlags:     get.NewGetPrintFlags(),
		IOStreams:      ioStreams,
	}

	cmd := &cobra.Command{
		Use:                   "pull [(--source-type SOURCE_TYPE) (--source-name SOURCE_NAME)]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Pulls files from Dataservice Operator"),
		Long:                  output.ReplaceCommandStrings(pullLong),
		Example:               output.ReplaceCommandStrings(pullExample),
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

	return cmd
}

type exportPullOptions struct {
	rhmConfigFlags *config.ConfigFlags
	PrintFlags     *get.PrintFlags

	// flags
	sourceName, sourceType string

	//internal
	args      []string
	rawConfig clientapi.Config

	printer printers.Printer

	genericclioptions.IOStreams
	rhmRawConfig *datactlapi.Config

	sources.Factory
}

func (e *exportPullOptions) Complete(cmd *cobra.Command, args []string) error {
	e.args = args
	var err error
	e.rhmRawConfig, err = e.rhmConfigFlags.RawPersistentConfigLoader().RawConfig()
	if err != nil {
		return err
	}

	e.PrintFlags.NamePrintFlags.Operation = "pull"

	e.printer, err = printers.NewPrinter(e.Out, e.PrintFlags)

	if err != nil {
		return err
	}

	e.Factory = (&sources.SourceFactoryBuilder{}).
		SetConfigFlags(e.rhmConfigFlags).
		SetPrinter(e.printer).
		Build()

	return nil
}

func (e *exportPullOptions) Validate() error {
	return nil
}

func (e *exportPullOptions) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	currentMeteringExport, err := e.rhmConfigFlags.MeteringExport()
	if err != nil {
		return err
	}

	bundleFile, err := bundle.NewBundleFromExport(currentMeteringExport)
	if err != nil {
		return err
	}
	errs := []error{}

	e.printer.HumanOutput(func(ho *output.HumanOutput) *output.HumanOutput {
		p := ho
		p.WithDetails("exportFile", currentMeteringExport.FileName).Titlef("%s", i18n.T("pull started"))
		p = p.Sub()
		return p
	})

	for name := range e.rhmRawConfig.Sources {
		s := e.rhmRawConfig.Sources[name]
		source, err := e.Factory.FromSource(*s)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		e.printer.HumanOutput(func(ho *output.HumanOutput) *output.HumanOutput {
			p := ho
			p.WithDetails("source", s.Name, "type", s.Type).Titlef("%s", i18n.T("pull started for source"))
			return p
		})

		err = source.Pull(ctx, currentMeteringExport, bundleFile, sources.EmptyOptions())
		if err != nil {
			errs = append(errs, err)
		}

		e.printer.HumanOutput(func(ho *output.HumanOutput) *output.HumanOutput {
			p := ho
			p.WithDetails("source", s.Name, "type", s.Type).Infof(i18n.T("pull complete"))
			return p
		})
	}

	fileNames := map[string]interface{}{}

	for _, f := range currentMeteringExport.Files {
		fileNames[f.Name] = nil
	}

	err = bundleFile.Close()
	if err != nil {
		return err
	}

	err = bundleFile.Compact(fileNames)
	if err != nil {
		return err
	}

	if err := config.ModifyConfig(e.rhmConfigFlags.ConfigAccess(), *e.rhmRawConfig, true); err != nil {
		return err
	}

	return nil
}
