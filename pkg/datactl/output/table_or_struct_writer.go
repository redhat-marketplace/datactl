package output

import (
	"io"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/kubectl/pkg/cmd/get"
)

type TableOrStructPrinter struct {
	ColumnDefinitions []metav1.TableColumnDefinition
	Printer           printers.ResourcePrinter
	PrintFlags        *get.PrintFlags
	Operation         func() string
	ObjectToRow       func(obj runtime.Object) metav1.TableRow

	table *metav1.Table
}

func (t *TableOrStructPrinter) PrintObj(obj runtime.Object, w io.Writer) error {
	if t.PrintFlags.OutputFormat != nil && *t.PrintFlags.OutputFormat != "wide" {
		return t.Printer.PrintObj(obj, w)
	}

	if t.table == nil {
		t.table = &metav1.Table{
			ColumnDefinitions: t.ColumnDefinitions,
		}
	}

	t.table.Rows = []metav1.TableRow{
		t.ObjectToRow(obj),
	}

	return t.Printer.PrintObj(t.table, w)
}
