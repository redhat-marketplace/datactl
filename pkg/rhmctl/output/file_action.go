package output

import (
	dataservicev1 "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api/dataservice/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/kubectl/pkg/cmd/get"
)

func NewActionCLITableOrStruct(
	flags *get.PrintFlags,
	printer printers.ResourcePrinter,
) *TableOrStructPrinter {
	return &TableOrStructPrinter{
		PrintFlags: flags,
		ColumnDefinitions: []metav1.TableColumnDefinition{
			{
				Name:        "ID",
				Description: "id of the file",
			},
			{
				Name:        "Name",
				Description: "name of the file",
			},
			{
				Name:        "Size",
				Description: "size of the file",
			},
			{
				Name:        "Committed",
				Description: "file has been committed on dataservice",
			},
			{
				Name:        "Pushed",
				Description: "file has been pushed to metric api",
			},
			{
				Name:        "Action",
				Description: "action taken",
			},
		},
		Printer: printer,
		ObjectToRow: func(obj runtime.Object) metav1.TableRow {
			file := obj.(*dataservicev1.FileInfoCTLAction)
			return metav1.TableRow{
				Cells: []interface{}{
					file.Id, file.Name, file.Size, file.Committed, file.Pushed, file.Action,
				},
			}
		},
	}
}

func NewPushFileOnlyCLITableOrStruct(
	flags *get.PrintFlags,
	printer printers.ResourcePrinter,
) *TableOrStructPrinter {
	return &TableOrStructPrinter{
		PrintFlags: flags,
		ColumnDefinitions: []metav1.TableColumnDefinition{
			{
				Name:        "Name",
				Description: "name of the file",
			},
			{
				Name:        "Size",
				Description: "size of the file",
			},
			{
				Name:        "Pushed",
				Description: "file has been pushed to metric api",
			},
			{
				Name:        "Action",
				Description: "action taken",
			},
		},
		Printer: printer,
		ObjectToRow: func(obj runtime.Object) metav1.TableRow {
			file := obj.(*dataservicev1.FileInfoCTLAction)
			return metav1.TableRow{
				Cells: []interface{}{
					file.Name, file.Size, file.Pushed, file.Action,
				},
			}
		},
	}
}
