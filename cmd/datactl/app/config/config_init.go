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

package config

import (
	"fmt"
	"io"

	"emperror.dev/errors"
	"github.com/manifoldco/promptui"
	datactlapi "github.com/redhat-marketplace/datactl/pkg/datactl/api"
	"github.com/redhat-marketplace/datactl/pkg/datactl/config"
	"github.com/redhat-marketplace/datactl/pkg/printers/output"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	configInitLong = templates.LongDesc(i18n.T(`
		Configures the default config file ('{{ .defaultConfigFile }}') with details about the cluster.
		It will also prompt for the Upload API endpoint and secret if they are not provided by flags.`))

	configInitExample = templates.Examples(i18n.T(`
		# Initialize the config, prompting for API and Token values.
		{{ .cmd }} config init

		# Initialize the config and preset upload URL and secret. Will not prompt.
		{{ .cmd }} config init --api marketplace.redhat.com --token MY_TOKEN
`))
)

func NewCmdConfigInit(rhmFlags *config.ConfigFlags, f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := configInitOptions{
		rhmConfigFlags: rhmFlags,
		IOStreams:      streams,
	}

	cmd := &cobra.Command{
		Use:                   "init",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Initializes the config for Dataservice and API endpoints"),
		Long:                  output.ReplaceCommandStrings(configInitLong),
		Example:               output.ReplaceCommandStrings(configInitExample),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}
	cmd.Flags().StringVar(&o.apiEndpoint, "api", "", i18n.T("upload endpoint"))
	cmd.Flags().StringVar(&o.apiSecret, "token", "", i18n.T("upload api secret"))

	return cmd
}

type configInitOptions struct {
	rhmConfigFlags  *config.ConfigFlags
	rhmConfigAccess config.ConfigAccess

	args         []string
	rhmRawConfig *datactlapi.Config

	apiEndpoint string
	apiSecret   string

	genericclioptions.IOStreams
}

func (init *configInitOptions) Complete(cmd *cobra.Command, args []string) error {
	init.args = args

	var err error
	init.rhmRawConfig, err = init.rhmConfigFlags.RawPersistentConfigLoader().RawConfig()
	if err != nil {
		return errors.Wrap(err, "error getting rhm config")
	}

	init.rhmConfigAccess = init.rhmConfigFlags.ConfigAccess()
	return nil
}

func (init *configInitOptions) Validate() error {
	return nil
}

func (init *configInitOptions) runAPIEndpointPrompt() error {
	validate := func(input string) error {
		if len(input) < 3 {
			return errors.New("Upload API Endpoint must be a valid domain.")
		}

		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Upload API Endpoint",
		Validate: validate,
		Default:  "marketplace.redhat.com",
		Stdin:    io.NopCloser(init.In),
		Stdout:   NopWCloser(init.Out),
	}

	fmt.Fprint(init.Out, i18n.T("Please provide the desired Upload API endpoint.\n"))
	result, err := prompt.Run()

	if err != nil {
		return err
	}

	init.rhmRawConfig.MarketplaceEndpoint.Host = "https://" + result
	return nil
}

func (init *configInitOptions) runAPISecretPrompt() error {
	validate := func(input string) error {
		if len(input) < 10 {
			return errors.New("Secret is too short.")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Upload API Secret",
		Validate: validate,
		Mask:     '*',
		Stdin:    io.NopCloser(init.In),
		Stdout:   NopWCloser(init.Out),
	}

	fmt.Fprint(init.Out, i18n.T("Please provide the secret for the Upload API.\n"))
	result, err := prompt.Run()
	if err != nil {
		return err
	}

	init.rhmRawConfig.MarketplaceEndpoint.PullSecretData = result
	return nil
}

func (init *configInitOptions) setUploadHost() error {
	if init.apiEndpoint == "" {
		err := init.runAPIEndpointPrompt()
		if err != nil {
			return err
		}
		return nil
	}

	init.rhmRawConfig.MarketplaceEndpoint.Host = init.apiEndpoint

	return nil
}

func (init *configInitOptions) setUploadSecret() error {
	if init.apiSecret == "" {
		err := init.runAPISecretPrompt()
		if err != nil {
			return err
		}
		return nil
	}

	init.rhmRawConfig.MarketplaceEndpoint.PullSecretData = init.apiSecret
	return nil
}

func (init *configInitOptions) runUploadAPIPrompts() error {
	if err := init.setUploadHost(); err != nil {
		return err
	}

	if err := init.setUploadSecret(); err != nil {
		return err
	}

	return nil
}

func (init *configInitOptions) Run() error {
	err := init.runUploadAPIPrompts()
	if err != nil {
		return err
	}

	if err := config.ModifyConfig(init.rhmConfigAccess, *init.rhmRawConfig, true); err != nil {
		return errors.Wrap(err, "error modifying config")
	}

	return nil
}

func NopWCloser(w io.Writer) io.WriteCloser {
	return nopWCloser{w}
}

type nopWCloser struct {
	io.Writer
}

func (nopWCloser) Close() error { return nil }
