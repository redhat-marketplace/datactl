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

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/fatih/color"
	"github.com/spf13/pflag"
	"k8s.io/kubectl/pkg/util/i18n"
)

type Padding int

const DefaultInitialPadding = 0

const ExtraPadding = DefaultInitialPadding + 3

func AddFlags(pf *pflag.FlagSet) {
	pf.BoolVar(&color.NoColor, "no-color", false, i18n.T("no color on CLI output"))
}

func Print(padding Padding, title string) {
	defer func() {
		cli.Default.Padding = int(DefaultInitialPadding)
	}()
	cli.Default.Padding = int(padding)
	log.Infof(color.New(color.Bold).Sprint(title))
}

func EnableColor() {
	color.NoColor = false
}

func DisableColor() {
	color.NoColor = false
}

func SetOutput(w io.Writer) {
	log.SetHandler(cli.Default)
	cli.Default.Writer = w
	cli.Strings = [...]string{
		log.DebugLevel: "*",
		log.InfoLevel:  "*",
		log.WarnLevel:  "*",
		log.ErrorLevel: "x",
		log.FatalLevel: "x",
	}
}

type HumanOutput struct {
	padding Padding
	fields  log.Fields
}

func NewHumanOutput() *HumanOutput {
	return newHumanOutput(DefaultInitialPadding)
}

func newHumanOutput(padding Padding) *HumanOutput {
	return &HumanOutput{padding: padding, fields: log.Fields{}}
}

func (h HumanOutput) Sub() *HumanOutput {
	return newHumanOutput(h.padding + ExtraPadding)
}

func (h HumanOutput) Println(a ...interface{}) {
	fmt.Fprintln(cli.Default.Writer, a...)
}

func (h *HumanOutput) WithDetails(key string, value interface{}, fields ...interface{}) *HumanOutput {
	output := newHumanOutput(h.padding)
	output.fields[key] = value

	if len(fields) > 0 && len(fields)%2 == 0 {
		for i := 0; i < len(fields); i = i + 2 {
			key, value := fields[i], fields[i+1]

			if k, ok := key.(string); ok {
				output.fields[k] = value
			}
			if k, ok := key.(fmt.Stringer); ok {
				output.fields[k.String()] = value
			}
		}
	}
	return output
}

func (h HumanOutput) Titlef(format string, a ...interface{}) {
	cli.Default.Padding = int(h.padding)
	log.WithFields(h.fields).Infof(color.New(color.Bold).Sprintf(format, a...))
}

func (h HumanOutput) Infof(format string, a ...interface{}) {
	cli.Default.Padding = int(h.padding)
	log.WithFields(h.fields).Infof(format, a...)
}

func (h HumanOutput) Warnf(format string, a ...interface{}) {
	cli.Default.Padding = int(h.padding)
	log.WithFields(h.fields).Warnf(color.New(color.FgYellow, color.Bold).Sprintf(format, a...))
}

func (h HumanOutput) Errorf(err error, format string, a ...interface{}) {
	cli.Default.Padding = int(h.padding)
	log.WithFields(h.fields).WithError(err).Errorf(format, a...)
}

func (h HumanOutput) Fatalf(err error, format string, a ...interface{}) {
	cli.Default.Padding = int(h.padding)
	log.WithFields(h.fields).WithError(err).Fatalf(format, a...)
}
