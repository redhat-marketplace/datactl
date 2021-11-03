package metrics

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	imageLong = templates.LongDesc(i18n.T(``))

	imageExample = templates.Examples(`
		# Example
		rhmctl upload foo`)
)

type MetricsUploadOptions struct {
	resource.FilenameOptions

	PrintObj   printers.ResourcePrinterFunc
	Output     string
	PrintFlags *genericclioptions.PrintFlags
	genericclioptions.IOStreams
}

// NewImageOptions returns an initialized MetricsUploadOptions instance
func NewMetricsUploadOptions(streams genericclioptions.IOStreams) *MetricsUploadOptions {
	return &MetricsUploadOptions{
		PrintFlags: genericclioptions.NewPrintFlags("metrics upload").WithTypeSetter(scheme.Scheme),
		IOStreams:  streams,
	}
}

func NewCmdUpload(f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewMetricsUploadOptions(streams)

	cmd := &cobra.Command{
		Use:                   "upload FILE_NAME_1 ... FILE_NAME_N | stdin",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Upload metrics from a file or stdin"),
		Long:                  imageLong,
		Example:               imageExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f, cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}

	o.PrintFlags.AddFlags(cmd)
	//o.RecordFlags.AddFlags(cmd)

	usage := "identifying the resource to get from a server."
	cmdutil.AddFilenameOptionFlags(cmd, &o.FilenameOptions, usage)
	cmdutil.AddDryRunFlag(cmd)

	return cmd
}

// Complete completes all required options
func (o *MetricsUploadOptions) Complete(f cmdutil.Factory, cmd *cobra.Command, args []string) error {
	var err error

	// o.RecordFlags.Complete(cmd)
	// o.Recorder, err = o.RecordFlags.ToRecorder()
	// if err != nil {
	// 	return err
	// }

	// o.UpdatePodSpecForObject = polymorphichelpers.UpdatePodSpecForObjectFn
	// o.DryRunStrategy, err = cmdutil.GetDryRunStrategy(cmd)
	// if err != nil {
	// 	return err
	// }

	// dynamicClient, err := f.DynamicClient()
	// if err != nil {
	// 	return err
	// }
	o.Output = cmdutil.GetFlagString(cmd, "output")

	printer, err := o.PrintFlags.ToPrinter()
	if err != nil {
		return err
	}

	o.PrintObj = printer.PrintObj

	// cmdNamespace, enforceNamespace, err := f.ToRawKubeConfigLoader().Namespace()
	// if err != nil {
	// 	return err
	// }

	// builder := f.NewBuilder().
	// 	WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
	// 	LocalParam(o.Local).
	// 	ContinueOnError().
	// 	NamespaceParam(cmdNamespace).DefaultNamespace().
	// 	FilenameParam(enforceNamespace, &o.FilenameOptions).
	// 	Flatten()

	return nil
}

// Validate makes sure provided values in MetricsUploadOptions are valid
func (o *MetricsUploadOptions) Validate() error {
	return nil
}

// Run performs the execution of 'set image' sub command
func (o *MetricsUploadOptions) Run() error {
	// allErrs := []error{}
	// if o.Local || o.DryRunStrategy == cmdutil.DryRunClient {
	// if err := o.PrintObj(info.Object, o.Out); err != nil {
	// 	allErrs = append(allErrs, err)
	// }
	// continue

	// if o.DryRunStrategy == cmdutil.DryRunServer {
	// 	if err := o.DryRunVerifier.HasSupport(info.Mapping.GroupVersionKind); err != nil {
	// 		return err
	// 	}
	// }
	return nil
}
