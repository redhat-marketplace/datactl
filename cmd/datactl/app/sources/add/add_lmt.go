package add

import (
	"fmt"
	"io"
	"strconv"

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
	configInitLongLmt = templates.LongDesc(i18n.T(`
		Adds a source of the licence metric tool to the config file with details about the source & source type.
		The command will attempt to add the source name & source type of the specified source in config.`))

	configInitExampleLmt = templates.Examples(i18n.T(`
		# Initialize the source, using the host, port and token.
		{{ .cmd }} sources add lmt --host host.example.com --port 443 --token aklsjfaskljfaslj
`))
)

const (
	LMT  string = "LMT"
	ILMT string = "ILMT"
	ilmt string = "ilmt"
)

func NewCmdAddLmt(rhmFlags *config.ConfigFlags, f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := addLmtOptions{
		rhmConfigFlags: rhmFlags,
		IOStreams:      streams,
	}

	cmd := &cobra.Command{
		Use:                   "lmt",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Initializes the config for LMT source details"),
		Long:                  output.ReplaceCommandStrings(configInitLongLmt),
		Example:               output.ReplaceCommandStrings(configInitExampleLmt),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}

	return cmd
}

type addLmtOptions struct {
	rhmConfigFlags  *config.ConfigFlags
	rhmConfigAccess config.ConfigAccess

	rhmRawConfig *datactlapi.Config
	lmtConfig    *datactlapi.LMTEndpoint

	LmtSourceType string
	Host          string
	Port          int
	Token         string

	genericclioptions.IOStreams
}

func (init *addLmtOptions) Complete(cmd *cobra.Command, args []string) error {

	var err error

	init.rhmRawConfig, err = init.rhmConfigFlags.RawPersistentConfigLoader().RawConfig()
	if err != nil {
		return errors.Wrap(err, "error getting rhm config")
	}

	init.rhmConfigAccess = init.rhmConfigFlags.ConfigAccess()

	return nil
}

func (init *addLmtOptions) Validate() error {
	return nil
}

func (init *addLmtOptions) Run() error {

	promptLmtSourceType := promptui.Prompt{
		Label:  fmt.Sprintf(i18n.T("Input %s source type"), LMT),
		Stdin:  io.NopCloser(init.In),
		Stdout: NopWCloser(init.Out),
	}

	lmtSourceType, err := promptLmtSourceType.Run()
	if err != nil {
		return err
	}

	promptHost := promptui.Prompt{
		Label:  fmt.Sprintf(i18n.T("Input %s host name"), LMT),
		Stdin:  io.NopCloser(init.In),
		Stdout: NopWCloser(init.Out),
	}

	host, err := promptHost.Run()
	if err != nil {
		return err
	}

	promptPort := promptui.Prompt{
		Label:  fmt.Sprintf(i18n.T("Input %s port no"), LMT),
		Stdin:  io.NopCloser(init.In),
		Stdout: NopWCloser(init.Out),
	}
	port, err := promptPort.Run()
	if err != nil {
		return err
	}

	promptToken := promptui.Prompt{
		Label:  fmt.Sprintf(i18n.T("Input %s token"), LMT),
		Stdin:  io.NopCloser(init.In),
		Stdout: NopWCloser(init.Out),
	}
	token, err := promptToken.Run()
	if err != nil {
		return err
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		return err
	}

	errr := init.addSourceDtlsToConfig(lmtSourceType, host, portInt, token)
	if errr != nil {
		return errr
	}

	return nil

}

//function to add source details i.e. source name and type in config
func (init *addLmtOptions) addSourceDtlsToConfig(lmtSourceType string, host string, port int, token string) error {
	if init.rhmRawConfig.LMTEndpoints == nil {
		init.rhmRawConfig.LMTEndpoints = make(map[string]*datactlapi.LMTEndpoint)
	}

	if _, ok := init.rhmRawConfig.LMTEndpoints[host]; !ok {
		init.rhmRawConfig.LMTEndpoints[host] = &datactlapi.LMTEndpoint{
			LMTSourceType: lmtSourceType,
			Host:          host,
			Port:          port,
			Token:         token,
		}
	}

	init.lmtConfig = init.rhmRawConfig.LMTEndpoints[host]

	if init.rhmRawConfig.Sources == nil {
		init.rhmRawConfig.Sources = make(map[string]*datactlapi.Source)
	}

	lmtSrcT, err := lmtSrcType(lmtSourceType)
	if err != nil {
		return err
	}

	init.rhmRawConfig.Sources[init.lmtConfig.Host] = &datactlapi.Source{
		Name: init.lmtConfig.Host,
		Type: lmtSrcT,
	}

	if err := config.ModifyConfig(init.rhmConfigAccess, *init.rhmRawConfig, true); err != nil {
		return errors.Wrap(err, "error modifying config")
	}

	return nil
}

//return the corresponding enum SourceType based on the lmt source type inpur
func lmtSrcType(lmtSourceType string) (datactlapi.SourceType, error) {
	if lmtSourceType == ILMT || lmtSourceType == ilmt {
		return datactlapi.ILMT, nil
	}
	return "", errors.New("Source type not defined or may be defined in future")
}
