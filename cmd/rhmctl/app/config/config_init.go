package config

import (
	"context"
	"fmt"
	"io"
	"strings"

	"emperror.dev/errors"
	"github.com/manifoldco/promptui"
	rhmctlapi "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	"github.com/redhat-marketplace/rhmctl/pkg/rhmctl/config"
	"github.com/redhat-marketplace/rhmctl/pkg/rhmctl/output"
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
		Configures the default config file (~/.rhmctl/config) with details about the cluster.
		The command will attempt to resolve the Dataservice URL. It will also prompt for the
		Upload API endpoint and secret if they are not provided by flags.`))

	configInitExample = templates.Examples(i18n.T(`
		# Initialize the config, prompting for API and Token values.
		%[1]s config init

		# Initialize the config and set upload URL and secret. Will not prompt.
		%[1]s config init --api marketplace.redhat.com --token MY_TOKEN
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
		Long:                  fmt.Sprintf(configInitLong, output.CommandName()),
		Example:               fmt.Sprintf(configInitExample, output.CommandName()),
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
	rhmRawConfig *rhmctlapi.Config

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

func (init *configInitOptions) Run() error {
	// conf, err := init.configFlags.ToRawKubeConfigLoader().ClientConfig()
	// if err != nil {
	// 	return err
	// }

	// client, err := kubernetes.NewForConfig(conf)
	// if err != nil {
	// 	return err
	// }

	// endpoint, ok := init.rhmRawConfig.DataServiceEndpoints[*init.configFlags.ClusterName]

	// client.CoreV1().Secrets()
	//
	//

	kconf, err := init.rhmConfigFlags.KubectlConfig.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return err
	}

	kcontext, ok := kconf.Contexts[kconf.CurrentContext]
	if !ok {
		return errors.New("current context not defined")
	}

	if _, ok := init.rhmRawConfig.DataServiceEndpoints[kcontext.Cluster]; !ok {
		init.rhmRawConfig.DataServiceEndpoints[kcontext.Cluster] = &rhmctlapi.DataServiceEndpoint{
			ClusterName: kcontext.Cluster,
		}
	}

	dataService := init.rhmRawConfig.DataServiceEndpoints[kcontext.Cluster]

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
		Namespace("openshift-redhat-marketplace").
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

	if dataServiceHost != "" {
		dataService.URL = "https://" + dataServiceHost
	}

	if dataService.ServiceAccount == "" {
		dataService.ServiceAccount = "default"
	}

	if dataService.Namespace == "" {
		dataService.Namespace = "openshift-redhat-marketplace"
	}

	if init.apiEndpoint == "" {
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
	} else {
		if strings.HasPrefix(init.apiEndpoint, "https://") {
			init.rhmRawConfig.MarketplaceEndpoint.Host = init.apiEndpoint
		} else {
			init.rhmRawConfig.MarketplaceEndpoint.Host = "https://" + init.apiEndpoint
		}
	}

	if init.apiSecret == "" {
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
	} else {
		init.rhmRawConfig.MarketplaceEndpoint.PullSecretData = init.apiSecret
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
