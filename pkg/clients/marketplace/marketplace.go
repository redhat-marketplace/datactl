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

	MarketplaceConfig

	RoundTripperConfig *shared.RoundTripperConfig

	metricClient *marketplaceMetricClient
}

type Client interface {
	Metrics() MarketplaceMetrics
}

func NewClient(config MarketplaceConfig) Client {
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
