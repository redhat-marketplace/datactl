package app

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

// NewCmdAlpha creates a command that acts as an alternate root command for features in alpha
func NewCmdAlpha(f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alpha",
		Short: i18n.T("Commands for features in alpha"),
		Long:  templates.LongDesc(i18n.T("These commands correspond to alpha features that are not enabled by default.")),
	}
	// NewKubeletCommand() will hide the alpha command if it has no subcommands. Overriding
	// the help function ensures a reasonable message if someone types the hidden command anyway.
	if !cmd.HasAvailableSubCommands() {
		cmd.SetHelpFunc(func(*cobra.Command, []string) {
			cmd.Println(i18n.T("No alpha commands are available in this version of rhmctl"))
		})
	}

	return cmd
}
