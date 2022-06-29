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

package metering

import (
	"context"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-marketplace/datactl/pkg/datactl/config"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd/get"
)

var _ = Describe("export_pull", func() {
	var (
		sut *exportPullOptions
	)

	BeforeEach(func() {
		sut = &exportPullOptions{}
	})

	It("should ", func() {

		Expect(sut).ToNot(BeNil())
	})
})

var _ = Describe("export_pull_ilmt", func() {

	Context("test if transfornation to ISC event done successfully without any error when fetching usage details from ILMT source for same start end date", func() {
		It("success", func() {

			kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			o := exportPullOptions{
				rhmConfigFlags: config.NewConfigFlags(kubeConfigFlags),
				PrintFlags:     get.NewGetPrintFlags(),
				sourceName:     "demo.ilmt.ibmcloudsecurity.com",
				sourceType:     "ILMT",
				startDate:      "2022-06-17",
				endDate:        "2022-06-17",
			}

			o.rhmRawConfig, _ = o.rhmConfigFlags.RawPersistentConfigLoader().RawConfig()
			o.Complete(nil, nil)
			for name := range o.rhmRawConfig.Sources {
				s := o.rhmRawConfig.Sources[name]
				if s.Type.String() == o.sourceType {
					_, _, err := o.IlmtPullBase(s, ctx, nil, nil)
					Expect(err).To(Succeed())
				}
			}
		})
	})

	Context("test if expected no of ISC event is being returned when fetching usage details from ILMT source for same start end date", func() {
		It("success", func() {
			kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			o := exportPullOptions{
				rhmConfigFlags: config.NewConfigFlags(kubeConfigFlags),
				PrintFlags:     get.NewGetPrintFlags(),
				sourceName:     "demo.ilmt.ibmcloudsecurity.com",
				sourceType:     "ILMT",
				startDate:      "2022-06-17",
				endDate:        "2022-06-17",
			}

			o.rhmRawConfig, _ = o.rhmConfigFlags.RawPersistentConfigLoader().RawConfig()
			o.Complete(nil, nil)
			for name := range o.rhmRawConfig.Sources {
				s := o.rhmRawConfig.Sources[name]
				if s.Type.String() == o.sourceType {
					count, _, _ := o.IlmtPullBase(s, ctx, nil, nil)
					Expect(1).To(Equal(count))
				}
			}
		})
	})

	Context("test if expected ISC event response is being returned when fetching usage details from ILMT source for same start end date", func() {
		It("success", func() {
			kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			o := exportPullOptions{
				rhmConfigFlags: config.NewConfigFlags(kubeConfigFlags),
				PrintFlags:     get.NewGetPrintFlags(),
				sourceName:     "demo.ilmt.ibmcloudsecurity.com",
				sourceType:     "ILMT",
				startDate:      "2022-06-17",
				endDate:        "2022-06-17",
			}

			o.rhmRawConfig, _ = o.rhmConfigFlags.RawPersistentConfigLoader().RawConfig()
			o.Complete(nil, nil)
			for name := range o.rhmRawConfig.Sources {
				s := o.rhmRawConfig.Sources[name]
				if s.Type.String() == o.sourceType {
					_, response, _ := o.IlmtPullBase(s, ctx, nil, nil)
					expectedRespPathComplete := os.Getenv("HOME")
					expectedRespPath := "/go/src/datactl/pkg/mocks/productusageresponse.json"
					expectedRespPathComplete += expectedRespPath
					expectedRespData, _ := os.ReadFile(expectedRespPathComplete)
					expectedRespDataStr := string(expectedRespData)
					Expect(expectedRespDataStr).To(Equal(response))
				}
			}
		})
	})
})
