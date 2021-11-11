package marketplace

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/redhat-marketplace/rhmctl/pkg/clients/shared"
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

func NewClient(config *MarketplaceConfig) Client {
	if config.polling == 0 {
		config.polling = 5 * time.Second
	}
	if config.timeout == 0 {
		config.timeout = 60 * time.Second
	}

	cli := &marketplaceClient{
		Client: shared.NewHttpClient(
			config.TlsConfig,
			shared.WithBearerAuth(config.Token),
		),
		MarketplaceConfig: config,
	}

	cli.metricClient = &marketplaceMetricClient{client: cli}
	return cli
}
