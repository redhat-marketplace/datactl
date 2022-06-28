package add

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/redhat-marketplace/datactl/pkg/datactl/config"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2/klogr"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
)

var (
	logger logr.Logger = klogr.New().V(5).WithName("sources")
)

func NewCmdAdd(rhmFlags *config.ConfigFlags, f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "add SUBCOMMAND",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Add a datactl source."),
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(NewCmdAddDataService(rhmFlags, f, streams))
	cmd.AddCommand(NewCmdAddIlmt(rhmFlags, f, streams))
	return cmd
}

func helpErrorf(cmd *cobra.Command, format string, args ...interface{}) error {
	cmd.Help()
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s", msg)
}
