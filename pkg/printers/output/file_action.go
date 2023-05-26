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

package output

import (
	"fmt"
	"io"

	dataservicev1 "github.com/redhat-marketplace/datactl/pkg/datactl/api/dataservice/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/kubectl/pkg/cmd/get"
)

func NewActionCLITableOrStruct(
	out io.Writer,
	flags *get.PrintFlags,
	printer printers.ResourcePrinter,
) *TableOrStructPrinter {
	writer := printers.GetNewTabWriter(out)
	return &TableOrStructPrinter{
		PrintFlags: flags,
		ColumnDefinitions: []metav1.TableColumnDefinition{
			{
				Name:        "      ID",
				Description: "id of the file",
			},
			{
				Name:        "Name                                                         ",
				Description: "name of the file",
				Type:        "string",
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
			{
				Name:        "Result",
				Description: "result of action",
			},
		},
		Printer: printer,
		ObjectToRow: func(obj runtime.Object) metav1.TableRow {
			file := obj.(*dataservicev1.FileInfoCTLAction)
			return metav1.TableRow{
				Cells: []interface{}{
					fmt.Sprintf("      %s", file.Id), file.Name, file.Size, file.Committed, file.Pushed, file.Action, file.Result,
				},
			}
		},
		w: writer,
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
				Name:        "      Name                                                         ",
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
			{
				Name:        "Result",
				Description: "result of action",
			},
		},
		Printer: printer,
		ObjectToRow: func(obj runtime.Object) metav1.TableRow {
			file := obj.(*dataservicev1.FileInfoCTLAction)
			return metav1.TableRow{
				Cells: []interface{}{
					fmt.Sprintf("     %s", file.Name), file.Size, file.Pushed, file.Action, file.Result,
				},
			}
		},
	}
}
