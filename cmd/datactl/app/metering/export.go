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
	"fmt"

	"github.com/go-logr/logr"
	"github.com/redhat-marketplace/datactl/pkg/datactl/config"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2/klogr"
	"k8s.io/kubectl/pkg/util/i18n"

	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

var (
	logger logr.Logger = klogr.New().V(5).WithName("export")
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
	cmd.AddCommand(NewCmdExportPush(rhmFlags, f, ioStreams))

	return cmd
}

// Idea is we have 1 active export
// 	- Tracked by a file, name can be given, default one is uniquely generated.
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
// datactl metering export start
// > Export started with
// datactl

func helpErrorf(cmd *cobra.Command, format string, args ...interface{}) error {
	cmd.Help()
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s", msg)
}
