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

//go:build tools
// +build tools

package main

import (
	_ "github.com/google/addlicense@latest"
	_ "github.com/google/go-licenses@v2"
	_ "github.com/onsi/ginkgo/ginkgo"
	_ "github.com/spf13/cobra/doc"
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen@v0.7.0"
)
