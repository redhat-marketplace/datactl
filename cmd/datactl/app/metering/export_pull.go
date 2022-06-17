package metering

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/redhat-marketplace/datactl/pkg/bundle"
	"github.com/redhat-marketplace/datactl/pkg/datactl/api"
	datactlapi "github.com/redhat-marketplace/datactl/pkg/datactl/api"
	"github.com/redhat-marketplace/datactl/pkg/datactl/config"
	"github.com/redhat-marketplace/datactl/pkg/printers"
	"github.com/redhat-marketplace/datactl/pkg/printers/output"
	"github.com/redhat-marketplace/datactl/pkg/sources"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	clientapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/kubectl/pkg/cmd/get"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	pullLong = templates.LongDesc(i18n.T(`
		Pulls data from all available sources. Filtering by source name and type is available.

		Prints a table of the files pulled with basic information.

		Please use the sources commands to add new sources for pulling.`))

	pullExample = templates.Examples(i18n.T(`
		# Pull all available data from all available sources and will prompt for start date in case of pull from ILMT
		{{ .cmd }} export pull all

		# Pull all data from a particular source-type. source-type flag is optional, if not given will pull for all the sources.
		{{ .cmd }} export pull all --source-type dataService/ilmt

		# Pull all data from a particular source. source-name flag is optional, if not given will pull for all the sources
		{{ .cmd }} export pull all --source-name my-dataservice-cluster/my-ilmt-server-hostname

		# Pull all data from a particular source and source type. source-type & source-name flags are optional, if not given will pull for all the sources
		{{ .cmd }} export pull all -source-type dataService/ilmt --source-name my-dataservice-cluster/my-ilmt-server-hostname

		# Pull all data from a particular source and source type. startdate and enddate flags are optional, if startdate, enddate not given for ILMT source will asks for prompt.
		{{ .cmd }} export pull all -source-type dataService/ilmt --source-name my-dataservice-cluster/my-ilmt-server-hostname --start-date 2022-02-04 --end-date 2022-06-02
`))
)

const (
	ILMT        string = "ILMT"
	DATASERVICE string = "DataService"
	EMPTY       string = ""
	StartDate          = "startDate"
	EndDate            = "endDate"
)

func NewCmdExportPull(rhmFlags *config.ConfigFlags, f cmdutil.Factory, ioStreams genericclioptions.IOStreams) *cobra.Command {
	o := exportPullOptions{
		rhmConfigFlags: rhmFlags,
		PrintFlags:     get.NewGetPrintFlags(),
		IOStreams:      ioStreams,
	}

	cmd := &cobra.Command{
		Use:                   "pull all [(--source-type SOURCE_TYPE) (--source-name SOURCE_NAME) (--startdate STARTDATE) (--enddate ENDDATE)]",
		DisableFlagsInUseLine: true,
		Short:                 i18n.T("Pulls files from Dataservice Operator/IBM Licence Metric Tool"),
		Long:                  output.ReplaceCommandStrings(pullLong),
		Example:               output.ReplaceCommandStrings(pullExample),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(cmd, args))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run())
		},
	}

	o.PrintFlags.AddFlags(cmd)

	cmd.Flags().StringVar(&o.sourceType, "source-type", EMPTY, i18n.T("Source Name"))
	cmd.Flags().StringVar(&o.sourceName, "source-name", EMPTY, i18n.T("Source Type"))
	cmd.Flags().StringVar(&o.startDate, "start-date", EMPTY, i18n.T("Start Date"))
	cmd.Flags().StringVar(&o.endDate, "end-date", EMPTY, i18n.T("End Date"))

	cmd.Flags().MarkHidden("label-columns")
	cmd.Flags().MarkHidden("sort-by")
	cmd.Flags().MarkHidden("show-kind")
	cmd.Flags().MarkHidden("show-managed-fields")
	cmd.Flags().MarkHidden("show-labels")

	return cmd
}

type exportPullOptions struct {
	rhmConfigFlags *config.ConfigFlags
	PrintFlags     *get.PrintFlags

	// flags
	sourceName, sourceType string

	//start & end date
	startDate, endDate string

	//internal
	args      []string
	rawConfig clientapi.Config

	printer printers.Printer

	genericclioptions.IOStreams
	rhmRawConfig *datactlapi.Config

	sources.Factory
}

func (e *exportPullOptions) Complete(cmd *cobra.Command, args []string) error {
	e.args = args
	var err error
	e.rhmRawConfig, err = e.rhmConfigFlags.RawPersistentConfigLoader().RawConfig()
	if err != nil {
		return err
	}

	e.PrintFlags.NamePrintFlags.Operation = "pull"

	e.printer, err = printers.NewPrinter(e.Out, e.PrintFlags)

	if err != nil {
		return err
	}

	e.Factory = (&sources.SourceFactoryBuilder{}).
		SetConfigFlags(e.rhmConfigFlags).
		SetPrinter(e.printer).
		Build()

	return nil
}

func (e *exportPullOptions) Validate() error {
	for name := range e.rhmRawConfig.Sources {
		s := e.rhmRawConfig.Sources[name]
		if s.Type == "ILMT" {
			if e.startDate == EMPTY {
				if e.rhmRawConfig.ILMTEndpoints[s.Name].LastPulldate == EMPTY {
					startDate, err := e.promptStartDate()
					if err != nil {
						e.printer.HumanOutput(func(ho *output.HumanOutput) *output.HumanOutput {
							p := ho
							p.Errorf(err, i18n.T(err.Error()))
							return p
						})
						os.Exit(1)
					}
					if startDate == EMPTY {
						e.printer.HumanOutput(func(ho *output.HumanOutput) *output.HumanOutput {
							p := ho
							p.Infof(i18n.T("Startdate mandatory to provide in case of pulling data from source first time"))
							return p
						})
						os.Exit(1)
					}
					e.startDate = startDate
				} else {
					e.startDate = e.rhmRawConfig.ILMTEndpoints[s.Name].LastPulldate
				}
			}

			re := regexp.MustCompile(`((19|20)\d\d)-(0?[1-9]|1[012])-(0?[1-9]|[12][0-9]|3[01])`)
			startDateMatched := re.MatchString(e.startDate)
			if !startDateMatched {
				e.printer.HumanOutput(func(ho *output.HumanOutput) *output.HumanOutput {
					p := ho
					p.Infof(i18n.T("Startdate must be in format yyyy-mm-dd"))
					return p
				})
				os.Exit(1)
			}

			startDate, err := time.Parse("2006-01-02", e.startDate)
			if err != nil {
				e.printer.HumanOutput(func(ho *output.HumanOutput) *output.HumanOutput {
					p := ho
					p.Errorf(err, i18n.T(err.Error()))
					return p
				})
				os.Exit(1)
			}
			isStartDateCheckFailed := startDate.After(time.Now().AddDate(0, 0, -1))
			if isStartDateCheckFailed {
				e.printer.HumanOutput(func(ho *output.HumanOutput) *output.HumanOutput {
					p := ho
					p.Infof(i18n.T("Start date must not be greater than yesterday date"))
					return p
				})
				os.Exit(1)
			}

			if e.endDate == EMPTY {
				yesterdayDate := fmt.Sprintf("%04d-%02d-%02d", time.Now().AddDate(0, 0, -1).Year(), time.Now().AddDate(0, 0, -1).Month(), time.Now().AddDate(0, 0, -1).Day())
				e.endDate = yesterdayDate
			}

			endDateMatched := re.MatchString(e.endDate)
			if !endDateMatched {
				e.printer.HumanOutput(func(ho *output.HumanOutput) *output.HumanOutput {
					p := ho
					p.Infof(i18n.T("Enddate must be in format yyyy-mm-dd"))
					return p
				})
				os.Exit(1)
			}

			endDate, err := time.Parse("2006-01-02", e.endDate)
			if err != nil {
				e.printer.HumanOutput(func(ho *output.HumanOutput) *output.HumanOutput {
					p := ho
					p.Errorf(err, i18n.T(err.Error()))
					return p
				})
				os.Exit(1)
			}
			isEndDateCheckFailed := endDate.After(time.Now().AddDate(0, 0, -1)) || endDate.Before(startDate)
			if isEndDateCheckFailed {
				e.printer.HumanOutput(func(ho *output.HumanOutput) *output.HumanOutput {
					p := ho
					p.Infof(i18n.T("End date must not be less than start date or greater than yesterday date"))
					return p
				})
				os.Exit(1)
			}

		} else if s.Type == "DataService" {
			continue
		}
		break
	}

	return nil
}

func (e *exportPullOptions) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	currentMeteringExport, err := e.rhmConfigFlags.MeteringExport()
	if err != nil {
		return err
	}

	bundleFile, err := bundle.NewBundleFromExport(currentMeteringExport)
	if err != nil {
		return err
	}

	for name := range e.rhmRawConfig.Sources {
		s := e.rhmRawConfig.Sources[name]

		if ((e.sourceType == EMPTY && e.sourceName == EMPTY) || (strings.EqualFold(e.sourceType, s.Type.String()) || strings.EqualFold(e.sourceName, s.Name))) && (strings.EqualFold(s.Type.String(), DATASERVICE)) {
			err := e.DataServicePullBase(s, ctx, currentMeteringExport, bundleFile)
			if err != nil {
				continue
			}
		} else if ((e.sourceType == EMPTY && e.sourceName == EMPTY) || (strings.EqualFold(e.sourceType, s.Type.String()) || strings.EqualFold(e.sourceName, s.Name))) && (strings.EqualFold(s.Type.String(), ILMT)) {
			_, _, err := e.IlmtPullBase(s, ctx, currentMeteringExport, bundleFile)
			if err != nil {
				continue
			}
			e.rhmRawConfig.ILMTEndpoints[s.Name].LastPulldate = strings.Split(time.Now().String(), " ")[0]
		}
	}

	fileNames := map[string]interface{}{}

	for _, f := range currentMeteringExport.Files {
		fileNames[f.Name] = nil
	}

	err = bundleFile.Close()
	if err != nil {
		return err
	}

	err = bundleFile.Compact(fileNames)
	if err != nil {
		return err
	}

	if err := config.ModifyConfig(e.rhmConfigFlags.ConfigAccess(), *e.rhmRawConfig, true); err != nil {
		return err
	}
	return nil
}

func (e *exportPullOptions) DataServicePullBase(s *datactlapi.Source, ctx context.Context,
	currentMeteringExport *api.MeteringExport,
	bundleFile *bundle.BundleFile) error {

	e.printer.HumanOutput(func(p *output.HumanOutput) *output.HumanOutput {
		p.WithDetails("exportFile", currentMeteringExport.FileName).Titlef(i18n.T("pulling sources to file"))
		return p.Sub()
	})

	source, err := e.Factory.FromSource(*s)
	if err != nil {
		e.printer.HumanOutput(func(ho *output.HumanOutput) *output.HumanOutput {
			p := ho
			p.Errorf(err, i18n.T("failed to get source"))
			return p
		})

		return err
	}

	e.printer.HumanOutput(func(p *output.HumanOutput) *output.HumanOutput {
		p = p.WithDetails("sourceName", s.Name, "sourceType", s.Type)
		p.Infof(i18n.T("pull start"))
		return p
	})

	count, err := source.Pull(ctx, currentMeteringExport, bundleFile, sources.EmptyOptions())

	if err != nil {
		e.printer.HumanOutput(func(p *output.HumanOutput) *output.HumanOutput {
			p.Errorf(err, i18n.T("pull failed"))
			return p
		})

		return err
	}

	e.printer.HumanOutput(func(p *output.HumanOutput) *output.HumanOutput {
		p.WithDetails("count", count).Infof(i18n.T("pull complete"))
		return p
	})

	return nil
}

func (e *exportPullOptions) IlmtPullBase(s *datactlapi.Source, ctx context.Context,
	currentMeteringExport *api.MeteringExport,
	bundleFile *bundle.BundleFile) (int, string, error) {
	source, err := e.Factory.FromSource(*s)
	if err != nil {
		e.printer.HumanOutput(func(ho *output.HumanOutput) *output.HumanOutput {
			p := ho
			p.Errorf(err, i18n.T("failed to get source"))
			return p
		})
		return -1, EMPTY, err
	}

	e.printer.HumanOutput(func(p *output.HumanOutput) *output.HumanOutput {
		p = p.WithDetails("sourceName", s.Name, "sourceType", s.Type)
		p.Infof(i18n.T("pull start"))
		return p
	})

	productCount, err := source.Pull(ctx, currentMeteringExport, bundleFile, sources.NewOptions(
		StartDate, e.startDate,
		EndDate, e.endDate,
	))

	productUsageResponseStr := source.GetResponse()
	fmt.Println(productUsageResponseStr)

	if err != nil {
		e.printer.HumanOutput(func(p *output.HumanOutput) *output.HumanOutput {
			p.Errorf(err, i18n.T("pull failed"))
			return p
		})

		return -1, EMPTY, err
	}

	e.printer.HumanOutput(func(p *output.HumanOutput) *output.HumanOutput {
		p.WithDetails("count", productCount).Infof(i18n.T("pull complete"))
		return p
	})

	return productCount, productUsageResponseStr, nil
}

func (e *exportPullOptions) promptStartDate() (string, error) {
	promptStartDate := promptui.Prompt{
		Label:  fmt.Sprintf(i18n.T("Enter start date in %s format"), "yyyy-mm-dd"),
		Stdin:  io.NopCloser(e.In),
		Stdout: NopWCloser(e.Out),
	}
	startDate, err := promptStartDate.Run()
	if err != nil {
		return EMPTY, err
	}
	return startDate, nil
}

func NopWCloser(w io.Writer) io.WriteCloser {
	return nopWCloser{w}
}

type nopWCloser struct {
	io.Writer
}

func (nopWCloser) Close() error { return nil }
