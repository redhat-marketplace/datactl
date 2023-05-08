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
	"io"

	"github.com/liggitt/tabwriter"
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
	w     *tabwriter.Writer
}

func (t *TableOrStructPrinter) Flush() {
	t.w.Flush()
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

func (t *TableOrStructPrinter) Print(obj runtime.Object) error {
	return t.PrintObj(obj, t.w)
}
