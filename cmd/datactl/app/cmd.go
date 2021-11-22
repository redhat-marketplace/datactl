package app

import (
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	configcmd "github.com/redhat-marketplace/datactl/cmd/datactl/app/config"
	"github.com/redhat-marketplace/datactl/cmd/datactl/app/metering"
	"github.com/redhat-marketplace/datactl/pkg/datactl/config"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	cliflag "k8s.io/component-base/cli/flag"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	"k8s.io/kubectl/pkg/util/term"
)

func NewDefaultDatactlCommand() *cobra.Command {
	return NewDatactlCommand(os.Stdin, os.Stdout, os.Stderr)
}

func NewDatactlCommand(in io.Reader, out, err io.Writer) *cobra.Command {
	warningHandler := rest.NewWarningWriter(err, rest.WarningWriterOptions{Deduplicate: true, Color: term.AllowsColorOutput(err)})
	warningsAsErrors := false
	// Parent command to which all subcommands are added.
	cmds := &cobra.Command{
		Use:   "datactl",
		Short: i18n.T("datactl controls the Red Hat Marketplace operator"),
		Long: templates.LongDesc(`
      datactl controls the Red Hat Marketplace operators.`),
		Run: runHelp,
		// Hook before and after Run initialize and write profiles to disk,
		// respectively.
		PersistentPreRunE: func(*cobra.Command, []string) error {
			rest.SetDefaultWarningHandler(warningHandler)
			return initProfiling()
		},
		PersistentPostRunE: func(*cobra.Command, []string) error {
			if err := flushProfiling(); err != nil {
				return err
			}
			if warningsAsErrors {
				count := warningHandler.WarningCount()
				switch count {
				case 0:
					// no warnings
				case 1:
					return fmt.Errorf("%d warning received", count)
				default:
					return fmt.Errorf("%d warnings received", count)
				}
			}
			return nil
		},
	}
	// From this point and forward we get warnings on flags that contain "_" separators
	// when adding them with hyphen instead of the original name.
	cmds.SetGlobalNormalizationFunc(cliflag.WarnWordSepNormalizeFunc)
	cmdutil.BehaviorOnFatal(func(msg string, exitCode int) {
		logrus.Fatalf(msg)
	})

	flags := cmds.PersistentFlags()

	addProfilingFlags(flags)

	flags.BoolVar(&warningsAsErrors, "warnings-as-errors", warningsAsErrors, "Treat warnings received from the server as errors and exit with a non-zero exit code")

	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	matchVersionKubeConfigFlags.AddFlags(flags)

	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	rhmConfigFlags := config.NewConfigFlags(kubeConfigFlags)
	rhmConfigFlags.AddFlags(flags)

	i18n.LoadTranslations("datactl", nil)

	ioStreams := genericclioptions.IOStreams{In: in, Out: out, ErrOut: err}

	groups := templates.CommandGroups{
		{
			Message: "Metering Commands:",
			Commands: []*cobra.Command{
				metering.NewCmdExport(rhmConfigFlags, f, ioStreams),
				//metering.NewCmdList(f, ioStreams),
			},
		},
		// {
		// 	Message:  "Troubleshooting and Debugging Commands:",
		// 	Commands: []*cobra.Command{
		// 		// Patch
		// 		// MustGather
		// 	},
		// },
		{
			Message: "Settings Commands:",
			Commands: []*cobra.Command{
				configcmd.NewCmdConfig(rhmConfigFlags, f, ioStreams),
				// Proxy
				// EnvVars
			},
		},
	}
	groups.Add(cmds)

	filters := []string{"options"}

	alpha := NewCmdAlpha(f, ioStreams)
	if !alpha.HasSubCommands() {
		filters = append(filters, alpha.Name())
	}

	templates.ActsAsRootCommand(cmds, filters, groups...)

	util.SetFactoryForCompletion(f)
	// registerCompletionFuncForGlobalFlags(cmds, f)

	//cmds.AddCommand(cmdconfig.NewCmdConfig(clientcmd.NewDefaultPathOptions(), ioStreams))

	cmds.SetGlobalNormalizationFunc(cliflag.WordSepNormalizeFunc)

	return cmds
}

func runHelp(cmd *cobra.Command, args []string) {
	cmd.Help()
}

// func registerCompletionFuncForGlobalFlags(cmd *cobra.Command, f cmdutil.Factory) {
// 	cmdutil.CheckErr(cmd.RegisterFlagCompletionFunc(
// 		"namespace",
// 		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
// 			return get.CompGetResource(f, cmd, "namespace", toComplete), cobra.ShellCompDirectiveNoFileComp
// 		}))
// 	cmdutil.CheckErr(cmd.RegisterFlagCompletionFunc(
// 		"context",
// 		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
// 			return util.ListContextsInConfig(toComplete), cobra.ShellCompDirectiveNoFileComp
// 		}))
// 	cmdutil.CheckErr(cmd.RegisterFlagCompletionFunc(
// 		"cluster",
// 		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
// 			return util.ListClustersInConfig(toComplete), cobra.ShellCompDirectiveNoFileComp
// 		}))
// 	cmdutil.CheckErr(cmd.RegisterFlagCompletionFunc(
// 		"user",
// 		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
// 			return util.ListUsersInConfig(toComplete), cobra.ShellCompDirectiveNoFileComp
// 		}))
// }