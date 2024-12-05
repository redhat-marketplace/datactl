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

package marketplace

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/redhat-marketplace/datactl/pkg/clients/shared"
)

type MarketplaceConfig struct {
	URL   string `json:"url"`
	Token string `json:"-"`

	TlsConfig *tls.Config

	polling time.Duration `json:"-"`
	timeout time.Duration `json:"-"`
}

type marketplaceClient struct {
	*http.Client

	*MarketplaceConfig

	RoundTripperConfig *shared.RoundTripperConfig

	metricClient *marketplaceMetricClient
}

type Client interface {
	Metrics() MarketplaceMetrics
}

func NewClient(config *MarketplaceConfig) (Client, error) {
	if config.polling == 0 {
		config.polling = 5 * time.Second
	}
	if config.timeout == 0 {
		config.timeout = 60 * time.Second
	}

	client, err := shared.NewHttpClient(
		config.TlsConfig,
		shared.WithBearerAuth(config.Token),
	)
	if err != nil {
		return nil, err
	}

	cli := &marketplaceClient{
		Client:            client,
		MarketplaceConfig: config,
	}

	cli.metricClient = &marketplaceMetricClient{client: cli}
	return cli, nil
}
