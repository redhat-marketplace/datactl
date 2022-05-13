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

package app

import (
	"fmt"
	"io"
	"os"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	configcmd "github.com/redhat-marketplace/datactl/cmd/datactl/app/config"
	"github.com/redhat-marketplace/datactl/cmd/datactl/app/metering"
	"github.com/redhat-marketplace/datactl/pkg/datactl/config"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2/klogr"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
	"k8s.io/kubectl/pkg/util/term"

	_ "embed"
	"github.com/redhat-marketplace/datactl/pkg/printers/output"
)

var (
	logger          logr.Logger = klogr.New()
	longDescription             = templates.LongDesc(i18n.T(`
		The datactl provides tooling for exporting data from disconnected
		cluster operators. To use this tool, there must be a Dataservice
		installed on the cluster and you must have access to it via the
		local network. The most optimal place is a jump host in a disconnected
		network.`))

	example = templates.Examples(i18n.T(`
		# Configure your system.
	 	{{ .cmd }} config init

		# Pull files from dataservice.
		{{ .cmd }} export pull

		# Commit files from dataservice; tells dataservice you
		# acknowledge their delivery.
		{{ .cmd }} export commit

		# Push to the configured upload API.
		{{ .cmd }} export push
`))
)

func NewDefaultDatactlCommand() *cobra.Command {
	return NewDatactlCommand(os.Stdin, os.Stdout, os.Stderr)
}

func NewDatactlCommand(in io.Reader, out, err io.Writer) *cobra.Command {
	warningHandler := rest.NewWarningWriter(err, rest.WarningWriterOptions{Deduplicate: true, Color: term.AllowsColorOutput(err)})
	warningsAsErrors := false

	// Parent command to which all subcommands are added.
	cmds := &cobra.Command{
		Use:     "datactl",
		Short:   i18n.T("datactl provides tooling to export data from operators"),
		Long:    longDescription,
		Example: output.ReplaceCommandStrings(example),
		Run:     runHelp,

		Version: version,
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

	cmds.SetVersionTemplate(`{{with .Name}}{{printf "%s " .}}{{end}}{{printf "%s" .Version}}
`)

	flags := cmds.PersistentFlags()

	cmds.PersistentFlags()
	addProfilingFlags(flags)

	flags.BoolVar(&warningsAsErrors, "warnings-as-errors", warningsAsErrors, "Treat warnings received from the server as errors and exit with a non-zero exit code")

	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	matchVersionKubeConfigFlags.AddFlags(flags)

	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	rhmConfigFlags := config.NewConfigFlags(kubeConfigFlags)
	rhmConfigFlags.AddFlags(flags)

	output.AddFlags(flags)

	i18n.LoadTranslations("datactl", nil)

	ioStreams := genericclioptions.IOStreams{In: in, Out: out, ErrOut: err}
	output.SetOutput(out)
	p := output.NewHumanOutput()

	cmdutil.BehaviorOnFatal(func(msg string, exitCode int) {
		p.Fatalf(nil, msg)
	})

	// From this point and forward we get warnings on flags that contain "_" separators
	// when adding them with hyphen instead of the original name.
	cmds.SetGlobalNormalizationFunc(cliflag.WarnWordSepNormalizeFunc)

	groups := templates.CommandGroups{
		{
			Message: "Metering Commands:",
			Commands: []*cobra.Command{
				metering.NewCmdExport(rhmConfigFlags, f, ioStreams),
			},
		},
		{
			Message: "Settings Commands:",
			Commands: []*cobra.Command{
				configcmd.NewCmdConfig(rhmConfigFlags, f, ioStreams),
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

	cmds.SetGlobalNormalizationFunc(cliflag.WordSepNormalizeFunc)

	return cmds
}

func runHelp(cmd *cobra.Command, args []string) {
	cmd.Help()
}
