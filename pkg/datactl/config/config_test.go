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

package config

import (
	"io/ioutil"
	"os"

	"github.com/gotidy/ptr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-marketplace/datactl/pkg/datactl/api"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var _ = Describe("config", func() {
	var (
		testConfig = `
marketplace:
  host: test.com
data-service-endpoints:
  - cluster-name: foo.test
    host: "foo.test"
    insecure-skip-tls-verify: true`

		f    *os.File
		name string
	)

	BeforeEach(func() {
		var err error
		tmpDir := os.TempDir()

		f, err = ioutil.TempFile(tmpDir, "config*.yaml")
		Expect(err).To(Succeed())

		_, err = f.WriteString(testConfig)
		Expect(err).To(Succeed())

		name = f.Name()

		Expect(f.Close()).To(Succeed())
	})

	AfterEach(func() {
		if f != nil {
			os.Remove(name)
		}
	})

	It("should parse v1", func() {
		conf, err := LoadFromFile(name)
		Expect(err).To(Succeed())
		Expect(conf).ToNot(BeNil())
		Expect(conf.DataServiceEndpoints).To(HaveLen(1))
		Expect(conf.DataServiceEndpoints["foo.test"]).ToNot(BeNil())
		Expect(conf.DataServiceEndpoints["foo.test"].Host).To(Equal("foo.test"))
	})

	It("should read file from flags", func() {
		testFlags := genericclioptions.NewConfigFlags(false)
		testFlags.Context = ptr.String("my-context")
		testFlags.ClusterName = ptr.String("foo")

		rhmConfigFlags := NewConfigFlags(testFlags)
		rhmConfigFlags.DATACTLConfig = ptr.String(name)

		conf, err := rhmConfigFlags.RawPersistentConfigLoader().RawConfig()
		Expect(err).To(Succeed())
		Expect(conf.DataServiceEndpoints).To(HaveLen(1))
		Expect(conf.DataServiceEndpoints["foo.test"]).ToNot(BeNil())
		Expect(conf.DataServiceEndpoints["foo.test"].Host).To(Equal("foo.test"))
	})

	It("should update file", func() {
		testFlags := genericclioptions.NewConfigFlags(false)
		testFlags.ClusterName = ptr.String("foo")
		testFlags.Context = ptr.String("my-context")

		rhmConfigFlags := NewConfigFlags(testFlags)
		rhmConfigFlags.DATACTLConfig = ptr.String(name)

		conf, err := rhmConfigFlags.RawPersistentConfigLoader().RawConfig()
		Expect(err).To(Succeed())

		conf.MeteringExports["foo"] = &api.MeteringExport{
			FileName:           "foo",
			DataServiceCluster: "foo.test",
		}

		Expect(ModifyConfig(rhmConfigFlags.ConfigAccess(), *conf, true)).To(Succeed())

		conf, err = LoadFromFile(name)
		Expect(err).To(Succeed())

		Expect(conf.DataServiceEndpoints).To(HaveLen(1))
		Expect(conf.DataServiceEndpoints["foo.test"]).ToNot(BeNil())
		Expect(conf.DataServiceEndpoints["foo.test"].Host).To(Equal("foo.test"))
		Expect(conf.MeteringExports).To(HaveLen(1))
	})
})
