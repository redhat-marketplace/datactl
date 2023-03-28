package add

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/gotidy/ptr"
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
		Adds a source of the dataservice type to the config file with details about the cluster.
		The command will attempt to resolve the Dataservice URL using the kubernetes context provided.`))

	configInitExample = templates.Examples(i18n.T(`
		# Initialize the source, using the default context.
		{{ .cmd }} sources add dataservice --use-default-context --namespace=redhat-marketplace

		# Initialize the source, using the default context and insecure self signed cert.
		{{ .cmd }} sources add dataservice --use-default-context --insecure-skip-tls-verify=true --allow-self-signed=true --namespace=redhat-marketplace

		# Initialize the config, prompting for a context to select.
		{{ .cmd }} sources add dataservice

		# Initialize the source, using the default context instead of selecting.
		{{ .cmd }} sources add --use-default-context
`))
)

func NewCmdAddDataService(rhmFlags *config.ConfigFlags, f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := addDataServiceOptions{
		rhmConfigFlags: rhmFlags,
		IOStreams:      streams,
	}

	cmd := &cobra.Command{
		Use:                   "dataservice",
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

	cmd.Flags().BoolVar(&o.useDefaultContext, "use-default-context", false, i18n.T("use the default kuberentes context instead of prompting"))
	cmd.Flags().BoolVar(&o.allowNonSystemCA, "allow-non-system-ca", false, i18n.T("allows non system CA certificates to be added to the dataService config"))
	cmd.Flags().BoolVar(&o.allowSelfsigned, "allow-self-signed", false, i18n.T("allows self-signed certificates to be added to the dataService configs"))

	return cmd
}

type addDataServiceOptions struct {
	rhmConfigFlags  *config.ConfigFlags
	rhmConfigAccess config.ConfigAccess

	args              []string
	rhmRawConfig      *datactlapi.Config
	dataServiceConfig *datactlapi.DataServiceEndpoint

	useDefaultContext bool
	allowNonSystemCA  bool
	allowSelfsigned   bool

	context string

	genericclioptions.IOStreams
}

func (init *addDataServiceOptions) Complete(cmd *cobra.Command, args []string) error {
	init.args = args

	var err error
	init.rhmRawConfig, err = init.rhmConfigFlags.RawPersistentConfigLoader().RawConfig()
	if err != nil {
		return errors.Wrap(err, "error getting rhm config")
	}

	init.rhmConfigAccess = init.rhmConfigFlags.ConfigAccess()

	return nil
}

func (init *addDataServiceOptions) Validate() error {
	return nil
}

func (init *addDataServiceOptions) runKubeConnected() error {
	// if not overridden, set them
	if init.dataServiceConfig.ServiceAccount == "" {
		init.dataServiceConfig.ServiceAccount = "default"
	}

	init.dataServiceConfig.Namespace = *init.rhmConfigFlags.KubectlConfig.Namespace

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

func (init *addDataServiceOptions) discoverDataServiceCA() error {
	// DataService cert uses serving-certs-ca-bundle, so the CA should already be in the pool
	// via the kube context. A user or test can still specify --insecure-skip-tls-verify
	conf := &tls.Config{
		InsecureSkipVerify: ptr.ToBool(init.rhmConfigFlags.KubectlConfig.Insecure),
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

func (init *addDataServiceOptions) runKubernetesContextPrompt() error {
	kconf, err := init.rhmConfigFlags.KubectlConfig.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return err
	}

	contexts := []string{}

	for name := range kconf.Contexts {
		contexts = append(contexts, name)
	}

	sort.Strings(contexts)

	defaultCursor := 0
	for i, name := range contexts {
		if name == kconf.CurrentContext {
			defaultCursor = i
			break
		}
	}

	prompt := promptui.Select{
		Label:     fmt.Sprintf(i18n.T("Select a %s context"), output.CommandName()),
		Items:     contexts,
		Stdin:     io.NopCloser(init.In),
		Stdout:    NopWCloser(init.Out),
		CursorPos: defaultCursor,
	}

	i, _, err := prompt.Run()
	if err != nil {
		return err
	}

	init.context = contexts[i]

	return nil
}

func (init *addDataServiceOptions) Run() error {
	kconf, err := init.rhmConfigFlags.KubectlConfig.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return err
	}

	if !init.useDefaultContext {
		if err := init.runKubernetesContextPrompt(); err != nil {
			return err
		}
	} else {
		init.context = kconf.CurrentContext
	}

	kcontext, ok := kconf.Contexts[init.context]

	if !ok {
		return errors.Errorf("%s context not defined", init.context)
	}

	if _, ok := init.rhmRawConfig.DataServiceEndpoints[init.context]; !ok {
		init.rhmRawConfig.DataServiceEndpoints[init.context] = &datactlapi.DataServiceEndpoint{
			ClusterName: kcontext.Cluster,
		}
	}

	init.dataServiceConfig = init.rhmRawConfig.DataServiceEndpoints[init.context]

	if err := init.runKubeConnected(); err != nil {
		return err
	}

	if init.rhmRawConfig.Sources == nil {
		init.rhmRawConfig.Sources = make(map[string]*datactlapi.Source)
	}

	init.rhmRawConfig.Sources[kcontext.Cluster] = &datactlapi.Source{
		Name: kcontext.Cluster,
		Type: datactlapi.DataService,
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
