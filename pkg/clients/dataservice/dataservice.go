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

package dataservice

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"github.com/redhat-marketplace/datactl/pkg/clients/shared"
	dataservicev1 "github.com/redhat-marketplace/datactl/pkg/datactl/api/dataservice/v1"
	"github.com/redhat-marketplace/datactl/pkg/datactl/api/latest"
	clientcmdlatest "github.com/redhat-marketplace/datactl/pkg/datactl/api/latest"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
)

var (
	logger logr.Logger = klogr.New().V(5).WithName("pkg/clients/dataservice")
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
	ListFiles(ctx context.Context, opts ListOptions, files *dataservicev1.ListFilesResponse) error
	GetFileById(ctx context.Context, id string, finfo *dataservicev1.FileInfo) (err error)
	GetFileByName(ctx context.Context, name, source, sourceType string, finfo *dataservicev1.FileInfo) (err error)
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

func (d *dataServiceClient) ListFiles(ctx context.Context, opts ListOptions, files *dataservicev1.ListFilesResponse) error {
	req, err := d.req.ListFiles(ctx)
	if err != nil {
		logger.Info("failed to get request", "err", err)
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

	filter := &strings.Builder{}

	if !opts.BeforeDate.IsZero() {
		filter.WriteString("createdAt < ")
		filter.WriteRune('"')
		filter.WriteString(opts.BeforeDate.Format(time.RFC3339))
		filter.WriteRune('"')
	}

	if !opts.AfterDate.IsZero() {
		if filter.Len() != 0 {
			filter.WriteString("&&")
		}
		filter.WriteString("createdAt > ")
		filter.WriteRune('"')
		filter.WriteString(opts.AfterDate.Format(time.RFC3339))
		filter.WriteRune('"')
	}

	if filter.Len() != 0 {
		q.Add("filter", filter.String())
	}

	req.URL.RawQuery = q.Encode()

	klog.V(5).Infof("url: %s, rawQuery: %s", req.URL.String(), " ", req.URL.RawQuery)

	resp, err := d.Do(req)
	if err != nil {
		klog.V(5).Infof("failed to get response: %v", err)
		return err
	}

	err = d.checkResponse("ListFiles", resp)
	if err != nil {
		return err
	}

	var body []byte

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		klog.V(5).Infof("failed to ready body %v", err)
	}

	listResponse := &dataservicev1.ListFilesResponse{}

	decoded, _, err := clientcmdlatest.Codec.Decode(body, &schema.GroupVersionKind{Version: latest.Version, Group: latest.Group, Kind: "ListFilesResponse"}, listResponse)
	if err != nil {
		klog.V(5).Infof("json decode failed err %v", err)
		return err
	}

	*files = *decoded.(*dataservicev1.ListFilesResponse)
	return nil
}

func (d *dataServiceClient) GetFileById(ctx context.Context, id string, finfo *dataservicev1.FileInfo) (err error) {
	req, err := d.req.GetFileByID(ctx, id)
	if err != nil {
		return err
	}
	return d.getFile(req, finfo)
}

func (d *dataServiceClient) GetFileByName(ctx context.Context, name, source, sourceType string, finfo *dataservicev1.FileInfo) (err error) {
	req, err := d.req.GetFileByName(ctx, name, source, sourceType)
	if err != nil {
		return err
	}

	return d.getFile(req, finfo)
}

func (d *dataServiceClient) DownloadFile(ctx context.Context, id string, w io.Writer) (checksum string, err error) {
	log := logger.WithValues("id", id)

	req, err := d.req.DownloadFile(ctx, id)
	if err != nil {
		log.Error(err, "failed to get request")
		return "", err
	}

	resp, err := d.Do(req)
	if err != nil {
		log.Error(err, "failed to run request")
		return "", err
	}

	err = d.checkResponse("DownloadFile", resp)
	if err != nil {
		log.Error(err, "response failed")
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

			log.Error(err, "failed to read file")
			return "", err
		}

		nAll = nAll + n
	}

	checksum = fmt.Sprintf("%x", sha.Sum(nil))
	return checksum, nil
}

func (d *dataServiceClient) getFile(req *http.Request, finfo *dataservicev1.FileInfo) (err error) {
	resp, err := d.Do(req)
	if err != nil {
		logger.Info("failed to get request", "err", err)
		return err
	}

	var body []byte
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		logger.Info("failed to read body", "err", err)
	}

	err = d.checkResponse("getFile", resp)
	if err != nil {
		return err
	}

	getResponse := &dataservicev1.GetFileResponse{}

	decoded, _, err := clientcmdlatest.Codec.Decode(body, &schema.GroupVersionKind{Version: clientcmdlatest.Version, Group: clientcmdlatest.Group, Kind: ""}, getResponse)
	if err != nil {
		logger.Info("failed to decode response", "err", err)
		return err
	}

	*finfo = *decoded.(*dataservicev1.GetFileResponse).Info
	return nil
}

func (d *dataServiceClient) DeleteFile(ctx context.Context, id string) error {
	log := logger.WithValues("id", id)

	req, err := d.req.DeleteFile(ctx, id)
	if err != nil {
		logger.Info("failed to get request", "err", err)
		return err
	}
	resp, err := d.Do(req)
	if err != nil {
		logger.Info("failed to get request", "err", err)
		return err
	}
	err = d.checkResponse("DeleteFile", resp)
	if err != nil {
		log.Error(err, "failed response", "statusCode", resp.StatusCode)
		return err
	}

	log.Info("deleted file")
	return nil
}

func (d *dataServiceClient) checkResponse(function string, resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		logger.Info("response",
			"func", function,
			"statusCode", resp.StatusCode,
			"body", string(body))
		return errors.NewWithDetails("failed request", "status", http.StatusText(resp.StatusCode), "body", strings.TrimSpace(string(body)))
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
