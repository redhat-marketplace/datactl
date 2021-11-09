package metering

import (
	"fmt"

	"github.com/redhat-marketplace/rhmctl/pkg/rhmctl/config"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/util/i18n"

	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func NewCmdExport(rhmFlags *config.ConfigFlags, f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "export",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Export metrics from RHM Operator"),
		// Long:                  imageLong,
		// Example:               imageExample,
		Run: func(cmd *cobra.Command, args []string) {
			// cmdutil.CheckErr(o.Complete(f, cmd, args))
			// cmdutil.CheckErr(o.Validate())
			// cmdutil.CheckErr(o.Run())
			cmd.Help()
		},
	}

	cmd.AddCommand(NewCmdExportPull(rhmFlags, f, ioStreams))
	cmd.AddCommand(NewCmdExportCommit(rhmFlags, f, ioStreams))
	cmd.AddCommand(NewCmdExportPush(f, ioStreams))

	return cmd
}

// Idea is we have 1 active export
// 	- Tracked by a file, name can be given, default one is uniquely generated.
// export start - sets a file as the active export and downloads to it.
//              - will error if there is already an active export
// export pull - pulls all the files to the active export
//              - will fail if there is not an active export
// export commit - marks the files on the server as deleted
//               - will fail if there is not an active export
// export push   - uploads the file to an available backend - can be run
//               - will list detected files and allow for one to be uploaded
//               - even though it is one file, each invididual payload is sent to the backend - serialy for first iteration.
//
// flow is:
//
// rhmctl metering export start
// > Export started with
// rhmctl

func helpErrorf(cmd *cobra.Command, format string, args ...interface{}) error {
	cmd.Help()
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s", msg)
}
