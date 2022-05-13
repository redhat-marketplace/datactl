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
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/manifoldco/promptui"
	datactlapi "github.com/redhat-marketplace/datactl/pkg/datactl/api"
	"github.com/redhat-marketplace/datactl/pkg/datactl/config"
	"github.com/redhat-marketplace/datactl/pkg/printers/output"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	configInitLong = templates.LongDesc(i18n.T(`
		Configures the default config file ('{{ .defaultConfigFile }}') with details about the cluster.
		The command will attempt to resolve the Dataservice URL. It will also prompt for the
		Upload API endpoint and secret if they are not provided by flags.

		If you are attempting to configure a host machine to upload payloads only; the --config-api-only
		flag is provided to prevent kubernetes resources from being queried. This is to prevent unnecessary
		errors with lack of access.`))

	configInitExample = templates.Examples(i18n.T(`
		# Initialize the config, prompting for API and Token values.
		{{ .cmd }} config init

		# Initialize the config and preset upload URL and secret. Will not prompt.
		{{ .cmd }} config init --api marketplace.redhat.com --token MY_TOKEN

		# Initialize only the API config, prompting for API and Token values.
		{{ .cmd }} config init --api-only

		# Initialize the config, force resetting of values if they are already set.
		{{ .cmd }} config init --force
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

	cmd.Flags().BoolVar(&o.force, "force", false, i18n.T("force configuration updates and prompts"))
	cmd.Flags().BoolVar(&o.apiOnly, "config-api-only", false, i18n.T("only configure Upload API components"))
	cmd.Flags().StringVar(&o.apiEndpoint, "api", "", i18n.T("upload endpoint"))
	cmd.Flags().StringVar(&o.apiSecret, "token", "", i18n.T("upload api secret"))
	cmd.Flags().BoolVar(&o.allowNonSystemCA, "allow-non-system-ca", false, i18n.T("allows non system CA certificates to be added to the dataService config"))
	cmd.Flags().BoolVar(&o.allowSelfsigned, "allow-self-signed", false, i18n.T("allows self-signed certificates to be added to the dataService configs"))

	return cmd
}

type configInitOptions struct {
	rhmConfigFlags  *config.ConfigFlags
	rhmConfigAccess config.ConfigAccess

	args              []string
	rhmRawConfig      *datactlapi.Config
	dataServiceConfig *datactlapi.DataServiceEndpoint

	apiEndpoint      string
	apiSecret        string
	apiOnly          bool
	force            bool
	allowNonSystemCA bool
	allowSelfsigned  bool

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

	if init.apiOnly {
		return nil
	}

	kconf, err := init.rhmConfigFlags.KubectlConfig.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return err
	}

	kcontext, ok := kconf.Contexts[kconf.CurrentContext]
	if !ok {
		return errors.New("current context not defined")
	}

	if _, ok := init.rhmRawConfig.DataServiceEndpoints[kcontext.Cluster]; !ok {
		init.rhmRawConfig.DataServiceEndpoints[kcontext.Cluster] = &datactlapi.DataServiceEndpoint{
			ClusterName: kcontext.Cluster,
		}
	}

	init.dataServiceConfig = init.rhmRawConfig.DataServiceEndpoints[kcontext.Cluster]
	return nil
}

func (init *configInitOptions) Validate() error {
	return nil
}

func (init *configInitOptions) runKubeConnected() error {
	// if not overridden, set them
	if init.dataServiceConfig.ServiceAccount == "" {
		init.dataServiceConfig.ServiceAccount = "default"
	}

	if init.dataServiceConfig.Namespace == "" {
		init.dataServiceConfig.Namespace = "openshift-redhat-marketplace"
	}

	restConfig, err := init.rhmConfigFlags.KubectlConfig.ToRESTConfig()
	if err != nil {
		return err
	}

	dynClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	// TODO namespace should be configured in the config file
	resource, err := dynClient.Resource(schema.GroupVersionResource{
		Group:    "route.openshift.io",
		Version:  "v1",
		Resource: "routes",
	}).
		Namespace(init.dataServiceConfig.Namespace).
		Get(context.TODO(), "rhm-data-service", metav1.GetOptions{})
	if err != nil {
		// check if we can't find route
		return err
	}

	var dataServiceHost string

	if spec, ok := resource.Object["spec"]; ok {
		if host, ok := spec.(map[string]interface{})["host"]; ok {
			dataServiceHost = host.(string)
		}
	}

	init.dataServiceConfig.Host = dataServiceHost

	if err := init.discoverDataServiceCA(); err != nil {
		return err
	}

	return nil
}

func (init *configInitOptions) discoverDataServiceCA() error {
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	pool, err := x509.SystemCertPool()
	if err != nil {
		pool = nil
	}

	containsSubject := func(cert *x509.Certificate) bool {
		if pool == nil {
			return false
		}

		for _, p := range pool.Subjects() {
			if bytes.Equal(p, cert.RawSubject) {
				return true
			}
		}

		return false
	}

	dataServiceHost := init.dataServiceConfig.Host
	if !strings.HasSuffix(dataServiceHost, ":443") && strings.IndexRune(dataServiceHost, ':') == -1 {
		dataServiceHost = dataServiceHost + ":443"
	}

	conn, err := tls.Dial("tcp", dataServiceHost, conf)
	if err != nil {
		logger.WithValues("err", err, "host", dataServiceHost).Info("failed to call dataservice")
		return fmt.Errorf("dataServiceHost is not reachable %s. Error: %s", dataServiceHost, err.Error())
	}
	defer conn.Close()
	certs := conn.ConnectionState().PeerCertificates
	for i, cert := range certs {
		expired := cert.NotAfter.Before(time.Now())

		if expired {
			return fmt.Errorf("certificate for dataServiceHost %s is expired", dataServiceHost)
		}

		if cert.Subject.String() == cert.Issuer.String() {
			if !init.allowSelfsigned {
				return fmt.Errorf("Certificate is self signed. To add to config file add %q flag to the command.", "--allow-self-signed")
			}

			logger.Info("certificate is self signed", "subject", cert.Subject, "issuer", cert.Issuer)
			init.dataServiceConfig.CertificateAuthorityData = cert.Raw
			break
		}

		if len(certs) == i+1 && cert.IsCA && containsSubject(cert) {
			logger.Info("cert pool contains subject so we don't have to store it")
			break
		}

		if len(certs) == i+1 && cert.IsCA {
			// found root CA
			if !init.allowNonSystemCA {
				return fmt.Errorf("Root CA not in system store. To add to config file add %q flag to the command.", "--allow-non-system-ca")
			}

			logger.Info(fmt.Sprintf("Found root CA but not in systemPool, adding to config file. subject: %s, issuer: %s", cert.Subject, cert.Issuer))
			init.dataServiceConfig.CertificateAuthorityData = cert.Raw
			break
		}
	}

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
	if init.rhmRawConfig.MarketplaceEndpoint.Host != "" && !init.force {
		return nil
	}

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
	if (init.rhmRawConfig.MarketplaceEndpoint.PullSecret != "" ||
		init.rhmRawConfig.MarketplaceEndpoint.PullSecretData != "") && !init.force {
		return nil
	}

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

func (init *configInitOptions) runOffline() error {
	return init.runUploadAPIPrompts()
}

func (init *configInitOptions) Run() error {
	err := init.runOffline()
	if err != nil {
		return err
	}

	if !init.apiOnly {
		err := init.runKubeConnected()
		if err != nil {
			return err
		}
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
