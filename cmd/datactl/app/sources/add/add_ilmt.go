package add

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
	configInitLongIlmt = templates.LongDesc(i18n.T(`
		The command will attempt to add the source name & type of the IBM Licence Metric Tool in config.`))

	configInitExampleIlmt = templates.Examples(i18n.T(`
		# Initialize the source, using the host, port and token.
		{{ .cmd }} sources add ilmt --host host.example.com --port 443 --token aklsjfaskljfaslj
`))
)

const (
	ILMT  string = "ILMT"
	EMPTY string = ""
)

func NewCmdAddIlmt(rhmFlags *config.ConfigFlags, f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := addIlmtOptions{
		rhmConfigFlags: rhmFlags,
		IOStreams:      streams,
	}

	cmd := &cobra.Command{
		Use:                   "ilmt",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Initializes the config for ILMT source details"),
		Long:                  output.ReplaceCommandStrings(configInitLongIlmt),
		Example:               output.ReplaceCommandStrings(configInitExampleIlmt),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}

	//adding flags for prompt
	cmd.Flags().StringVar(&o.Host, "Host", EMPTY, i18n.T("Host name of the ILMT source"))
	cmd.Flags().StringVar(&o.Port, "Port", EMPTY, i18n.T("Port number of the ILMT source"))
	cmd.Flags().StringVar(&o.Token, "Token", EMPTY, i18n.T("Token for accessing ILMT API"))

	return cmd
}

type addIlmtOptions struct {
	rhmConfigFlags  *config.ConfigFlags
	rhmConfigAccess config.ConfigAccess

	rhmRawConfig *datactlapi.Config
	lmtConfig    *datactlapi.ILMTEndpoint

	Host  string
	Port  string
	Token string

	genericclioptions.IOStreams
}

func (init *addIlmtOptions) Complete(cmd *cobra.Command, args []string) error {

	var err error

	init.rhmRawConfig, err = init.rhmConfigFlags.RawPersistentConfigLoader().RawConfig()
	if err != nil {
		return errors.Wrap(err, "error getting rhm config")
	}

	init.rhmConfigAccess = init.rhmConfigFlags.ConfigAccess()

	return nil
}

func (init *addIlmtOptions) Validate() error {
	return nil
}

func (init *addIlmtOptions) Run() error {

	if init.Host == EMPTY {
		if err := init.promptHost(); err != nil {
			return err
		}
	}

	if init.Port == EMPTY {
		if err := init.promptPort(); err != nil {
			return err
		}
	}

	if init.Token == EMPTY {
		if err := init.promptToken(); err != nil {
			return err
		}
	}

	err := init.addSourceDtlsToConfig(init.Host, init.Port, init.Token)
	if err != nil {
		return err
	}

	return nil

}

func (init *addIlmtOptions) promptHost() error {
	promptHost := promptui.Prompt{
		Label:  fmt.Sprintf(i18n.T("Enter %s host name"), ILMT),
		Stdin:  io.NopCloser(init.In),
		Stdout: NopWCloser(init.Out),
	}
	host, err := promptHost.Run()
	if err != nil {
		return err
	}
	if host == EMPTY {
		return errors.New("host name not provided for ILMT server")
	}
	init.Host = host
	return nil
}

func (init *addIlmtOptions) promptPort() error {
	promptPort := promptui.Prompt{
		Label:  fmt.Sprintf(i18n.T("Enter %s port number"), ILMT),
		Stdin:  io.NopCloser(init.In),
		Stdout: NopWCloser(init.Out),
	}
	port, err := promptPort.Run()
	if err != nil {
		return err
	}
	init.Port = port
	return nil
}

func (init *addIlmtOptions) promptToken() error {
	promptToken := promptui.Prompt{
		Label:  fmt.Sprintf(i18n.T("Enter %s token for access"), ILMT),
		Stdin:  io.NopCloser(init.In),
		Stdout: NopWCloser(init.Out),
	}
	token, err := promptToken.Run()
	if err != nil {
		return err
	}
	if token == EMPTY {
		return errors.New("token not provided for accessing ILMT API")
	}
	init.Token = token
	return nil
}

//function to add source details i.e. source name and type in config
func (init *addIlmtOptions) addSourceDtlsToConfig(host string, port string, token string) error {
	if init.rhmRawConfig.ILMTEndpoints == nil {
		init.rhmRawConfig.ILMTEndpoints = make(map[string]*datactlapi.ILMTEndpoint)
	}

	if _, ok := init.rhmRawConfig.ILMTEndpoints[host]; !ok {
		init.rhmRawConfig.ILMTEndpoints[host] = &datactlapi.ILMTEndpoint{
			Host:  host,
			Port:  port,
			Token: token,
		}
	}

	init.lmtConfig = init.rhmRawConfig.ILMTEndpoints[host]

	if init.rhmRawConfig.Sources == nil {
		init.rhmRawConfig.Sources = make(map[string]*datactlapi.Source)
	}

	init.rhmRawConfig.Sources[init.lmtConfig.Host] = &datactlapi.Source{
		Name: init.lmtConfig.Host,
		Type: datactlapi.ILMT,
	}

	if err := config.ModifyConfig(init.rhmConfigAccess, *init.rhmRawConfig, true); err != nil {
		return errors.Wrap(err, "error modifying config")
	}

	return nil
}
