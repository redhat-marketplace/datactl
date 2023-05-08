package printers

import (
	"io"

	"github.com/gotidy/ptr"
	"github.com/redhat-marketplace/datactl/pkg/printers/output"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubectl/pkg/cmd/get"
)

type PrintObj interface {
	Print(obj runtime.Object) error
}

type HumanPrinterCallback func(*output.HumanOutput) *output.HumanOutput
type TableOutputCallback func(PrintObj)

type TablePrinter interface {
	TableOutput(TableOutputCallback)
}

type HumanPrinter interface {
	HumanOutput(HumanPrinterCallback)
}

type Printer interface {
	TablePrinter
	HumanPrinter
}

type printWrapper struct {
	isHumanOutput bool
	h             *output.HumanOutput
	t             *output.TableOrStructPrinter
}

// HumanOutput provides a function with a printer. The printer can be saved for
// later calls by returning it. Or returning nil to reset it.
func (p *printWrapper) HumanOutput(print HumanPrinterCallback) {
	if p.isHumanOutput {
		h := print(p.h)
		if h != nil {
			p.h = h
		} else {
			p.h = output.NewHumanOutput()
		}
	}
}

func (p *printWrapper) TableOutput(print TableOutputCallback) {
	print(p.t)
	p.t.Flush()
}

func NewPrinter(
	out io.Writer,
	printFlags *get.PrintFlags,
) (Printer, error) {
	p := output.NewHumanOutput()
	print, err := printFlags.ToPrinter()
	if err != nil {
		return nil, err
	}

	humanOutput := false

	if printFlags.OutputFormat == nil ||
		*printFlags.OutputFormat == "wide" ||
		*printFlags.OutputFormat == "" {
		humanOutput = true
		printFlags.OutputFormat = ptr.String("wide")
	} else {
		output.DisableColor()
	}

	return &printWrapper{
		isHumanOutput: humanOutput,
		h:             p,
		t:             output.NewActionCLITableOrStruct(out, printFlags, print),
	}, nil
}
