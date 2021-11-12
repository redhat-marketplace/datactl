package dataservice

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"emperror.dev/errors"
	"github.com/redhat-marketplace/rhmctl/pkg/clients/shared"
	api "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	clientcmdapi "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api"
	"github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api/latest"
	clientcmdlatest "github.com/redhat-marketplace/rhmctl/pkg/rhmctl/api/latest"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"
)

type DataServiceConfig struct {
	URL   string `json:"url"`
	Token string `json:"-"`

	TlsConfig *tls.Config
}

type dataServiceClient struct {
	*http.Client

	RoundTripperConfig *shared.RoundTripperConfig
	DataServiceConfig
	req *reqBuilder
}

type Client interface {
	DownloadFile(ctx context.Context, id string, w io.Writer) (checksum string, err error)
	ListFiles(ctx context.Context, opts ListOptions, files *api.ListFilesResponse) error
	GetFileById(ctx context.Context, id string, finfo *api.FileInfo) (err error)
	GetFileByName(ctx context.Context, name, source, sourceType string, finfo *api.FileInfo) (err error)
	DeleteFile(context.Context, string) error
}

func NewClient(config *DataServiceConfig) Client {
	cli := &dataServiceClient{
		Client: shared.NewHttpClient(
			config.TlsConfig,
			shared.WithBearerAuth(config.Token),
		),
		DataServiceConfig: *config,
		req:               &reqBuilder{URL: config.URL},
	}
	return cli
}

const (
	queryPageToken      = "pageToken"
	queryPageSize       = "pageSize"
	queryIncludeDeleted = "includeDeleted"
)

type ListOptions struct {
	BeforeDate     time.Time
	AfterDate      time.Time
	OrderBy        string
	PageToken      string
	PageSize       *int
	IncludeDeleted bool
}

func (d *dataServiceClient) ListFiles(ctx context.Context, opts ListOptions, files *clientcmdapi.ListFilesResponse) error {
	req, err := d.req.ListFiles(ctx)
	if err != nil {
		logrus.WithError(err).Error("failed to get request")
		return err
	}

	q := req.URL.Query()

	if opts.PageToken != "" {
		q.Add(queryPageToken, opts.PageToken)
	}

	if opts.PageSize != nil {
		q.Add(queryPageSize, fmt.Sprintf("%d", opts.PageSize))
	}

	if opts.IncludeDeleted {
		q.Add(queryIncludeDeleted, "true")
	}

	req.URL.RawQuery = q.Encode()

	klog.Info("url is "+req.URL.String(), " ", req.URL.RawQuery)

	resp, err := d.Do(req)
	if err != nil {
		logrus.WithError(err).Error("failed to get response")
		return err
	}

	err = d.checkResponse("ListFiles", resp)
	if err != nil {
		logrus.WithField("statusCode", resp.StatusCode).
			WithError(err).
			Error("failed response")
		return err
	}

	var body []byte

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Error("failed to read body")
	}

	listResponse := &clientcmdapi.ListFilesResponse{}

	decoded, _, err := clientcmdlatest.Codec.Decode(body, &schema.GroupVersionKind{Version: latest.Version, Group: latest.Group, Kind: "ListFilesResponse"}, listResponse)
	if err != nil {
		logrus.WithError(err).Error("failed to decode response")
		return err
	}

	*files = *decoded.(*clientcmdapi.ListFilesResponse)
	return nil
}

func (d *dataServiceClient) GetFileById(ctx context.Context, id string, finfo *api.FileInfo) (err error) {
	req, err := d.req.GetFileByID(ctx, id)
	if err != nil {
		logrus.WithError(err).Error("failed to get request")
		return err
	}
	return d.getFile(req, finfo)
}

func (d *dataServiceClient) GetFileByName(ctx context.Context, name, source, sourceType string, finfo *api.FileInfo) (err error) {
	req, err := d.req.GetFileByName(ctx, name, source, sourceType)
	if err != nil {
		logrus.WithError(err).Error("failed to get request")
		return err
	}

	return d.getFile(req, finfo)
}

func (d *dataServiceClient) DownloadFile(ctx context.Context, id string, w io.Writer) (checksum string, err error) {
	log := logrus.WithField("id", id)

	req, err := d.req.DownloadFile(ctx, id)
	if err != nil {
		log.WithError(err).Error("failed to get request")
		return "", err
	}

	resp, err := d.Do(req)
	if err != nil {
		log.WithError(err).Error("failed to get request")
		return "", err
	}

	err = d.checkResponse("DownloadFile", resp)
	if err != nil {
		log.WithField("statusCode", resp.StatusCode).
			WithError(err).
			Error("failed response")
		return "", err
	}

	defer resp.Body.Close()

	sha := sha256.New()
	r := io.TeeReader(resp.Body, w)
	r = io.TeeReader(r, sha)

	// stream response
	p := make([]byte, 100)
	nAll := 0

	for {
		n, err := r.Read(p)
		if err != nil {
			if err == io.EOF {
				break
			}

			log.WithError(err).Error("failed to read file")
			return "", err
		}

		nAll = nAll + n
	}

	// TODO: add header check for checksum; should match
	checksum = fmt.Sprintf("%x", sha.Sum(nil))
	log.WithField("checksum", checksum).Debug("checksum calculated")

	return checksum, nil
}

func (d *dataServiceClient) getFile(req *http.Request, finfo *api.FileInfo) (err error) {
	resp, err := d.Do(req)
	if err != nil {
		logrus.WithError(err).Error("failed to get request")
		return err
	}

	var body []byte
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Error("failed to read body")
	}

	err = d.checkResponse("getFile", resp)
	if err != nil {
		return err
	}

	getResponse := &clientcmdapi.GetFileResponse{}

	decoded, _, err := clientcmdlatest.Codec.Decode(body, &schema.GroupVersionKind{Version: clientcmdlatest.Version, Group: clientcmdlatest.Group, Kind: ""}, getResponse)
	if err != nil {
		logrus.WithError(err).Error("failed to decode response")
		return err
	}

	*finfo = *decoded.(*clientcmdapi.GetFileResponse).Info
	return nil
}

func (d *dataServiceClient) DeleteFile(ctx context.Context, id string) error {
	log := logrus.WithField("id", id)

	req, err := d.req.DeleteFile(ctx, id)
	if err != nil {
		log.WithError(err).Error("failed to get request")
		return err
	}
	resp, err := d.Do(req)
	if err != nil {
		log.WithError(err).Error("failed to get request")
		return err
	}
	err = d.checkResponse("DeleteFile", resp)
	if err != nil {
		log.WithField("statusCode", resp.StatusCode).
			WithError(err).
			Error("failed response")
		return err
	}

	log.Info("deleted file")
	return nil
}

func (d *dataServiceClient) checkResponse(function string, resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		logrus.WithFields(logrus.Fields{
			"function":   function,
			"statusCode": resp.StatusCode,
			"body":       string(body),
		}).Error("failed response")
		return errors.NewWithDetails("failed request", "status", resp.StatusCode, "body", string(body))
	}

	return nil
}

type reqBuilder struct {
	URL string
}

func (u *reqBuilder) ListFiles(ctx context.Context) (*http.Request, error) {
	url := fmt.Sprintf("%s/v1/files", u.URL)
	return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
}

func (u *reqBuilder) GetFileByID(ctx context.Context, id string) (*http.Request, error) {
	url := fmt.Sprintf("%s/v1/files/%s", u.URL, id)
	return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
}

func (u *reqBuilder) GetFileByName(ctx context.Context, name, source, sourceType string) (*http.Request, error) {
	url := fmt.Sprintf("%s/v1/files/source/%s/sourceType/%s/name/%s", u.URL, source, sourceType, name)
	return http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
}

func (u *reqBuilder) DownloadFile(ctx context.Context, id string) (*http.Request, error) {
	url := fmt.Sprintf("%s/v1/file/%s/download", u.URL, id)
	return http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
}

func (u *reqBuilder) DeleteFile(ctx context.Context, id string) (*http.Request, error) {
	url := fmt.Sprintf("%s/v1/files/%s", u.URL, id)
	return http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
}
