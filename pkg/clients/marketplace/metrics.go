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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"path/filepath"
	"time"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2/klogr"
)

var (
	logger logr.Logger = klogr.New().V(5)
)

type MarketplaceUsageResponseDetails struct {
	Code       string `json:"code"`
	StatusCode int    `json:"statusCode"`
	Retryable  bool   `json:"retryable,omitempty"`
}

type MarketplaceUsageResponse struct {
	RequestID     string      `json:"requestId,omitempty"`
	CorrelationID string      `json:"correlationId,omitempty"`
	Status        MktplStatus `json:"status,omitempty"`
	Message       string      `json:"message,omitempty"`
	ErrorCode     string      `json:"errorCode,omitempty"`

	Details *MarketplaceUsageResponseDetails `json:"details,omitempty"`
}

type MktplStatus string

const (
	MktplStatusSuccess    MktplStatus = "success"
	MktplStatusInProgress MktplStatus = "inProgress"
	MktplStatusFailed     MktplStatus = "failed"

	marketplaceMetricsPath   = "%s/metering/api/v2/metrics"
	marketplaceMetricsStatus = marketplaceMetricsPath + "/%s"
)

func (s *marketplaceClient) Metrics() MarketplaceMetrics {
	return s.metricClient
}

type MarketplaceMetrics interface {
	Status(ctx context.Context, id string) (*MarketplaceUsageResponse, error)
	Upload(ctx context.Context, fileName string, reader io.Reader) (id string, err error)
}

type marketplaceMetricClient struct {
	client *marketplaceClient
}

func (r *marketplaceMetricClient) Status(ctx context.Context, id string) (*MarketplaceUsageResponse, error) {
	status := MarketplaceUsageResponse{}
	status.Details = &MarketplaceUsageResponseDetails{}

	url := fmt.Sprintf(marketplaceMetricsStatus, r.client.URL, id)
	resp, err := r.client.Get(url)
	if err != nil {
		return &status, err
	}

	status.Details.StatusCode = resp.StatusCode

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &status, err
	}

	jsonErr := json.Unmarshal(data, &status)

	if err := checkError(resp, string(data), status, "failed to get status"); err != nil {
		return &status, err
	}

	if jsonErr != nil {
		return &status, err
	}

	return &status, nil
}

const RetryableError = errors.Sentinel("retryable")

const DuplicateError = errors.Sentinel("duplicate")

func isRetryable(err error) bool {
	if errors.Is(err, RetryableError) {
		return true
	}
	return false
}

func checkError(resp *http.Response, body string, status MarketplaceUsageResponse, message string) error {
	logger.Info("retrieved response",
		"statusCode", resp.StatusCode,
		"proto", resp.Proto,
		"status", status,
		"headers", resp.Header,
		"body", body,
	)

	/*
		200 - Complete
		202 - In Progress
		4xx - Format Error
		503 - System Outage
	*/

	if resp.StatusCode < 300 && resp.StatusCode >= 200 {
		return nil
	}

	err := errors.NewWithDetails(message, "code", resp.StatusCode, "message", status.Message, "errorCode", status.ErrorCode)

	if resp.StatusCode == http.StatusConflict {
		return errors.WrapWithDetails(DuplicateError, "status", status.Message)
	}
	// return status says retrybale
	if status.Details != nil && status.Details.Retryable {
		return errors.WrapWithDetails(RetryableError, "retryable error", append(errors.GetDetails(err), "message", err.Error())...)
	}

	return err
}

func (r *marketplaceMetricClient) makeFormFile(fileName string, file []byte) (form []byte, formContent string, err error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fileBase := filepath.Base(fileName)
	fileExt := filepath.Ext(fileBase)
	fileWithoutExt := fileBase[:len(fileBase)-len(fileExt)]

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fileWithoutExt, fileBase))
	h.Set("Content-Type", "application/gzip")

	var part io.Writer
	part, err = writer.CreatePart(h)

	if err != nil {
		return
	}

	_, err = io.Copy(part, bytes.NewReader(file))
	if err != nil {
		return
	}

	err = writer.Close()
	if err != nil {
		return
	}

	form = body.Bytes()
	formContent = writer.FormDataContentType()
	err = nil
	return
}

func (r *marketplaceMetricClient) uploadFile(ctx context.Context, form []byte, formContent string) (id string, err error) {
	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, "POST", fmt.Sprintf(marketplaceMetricsPath, r.client.URL), bytes.NewReader(form))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", formContent)

	// Perform the request
	resp, err := r.client.Do(req)
	if err != nil {
		logger.Info("failed to post", "err", err)
		return "", errors.Wrap(err, "failed to post")
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "failed to read response body")
	}

	status := MarketplaceUsageResponse{}
	jsonErr := json.Unmarshal(body, &status)

	if err := checkError(resp, string(body), status, "failed to upload"); err != nil {
		return "", err
	}

	if jsonErr != nil {
		return "", err
	}

	return status.RequestID, err
}

var DefaultBackoff = wait.Backoff{
	Steps:    4,
	Duration: 50 * time.Millisecond,
	Factor:   5.0,
	Jitter:   0.1,
}

func (r *marketplaceMetricClient) Upload(ctx context.Context, fileName string, reader io.Reader) (id string, err error) {
	file, ferr := io.ReadAll(reader)
	if ferr != nil {
		err = ferr
		return
	}

	body, content, berr := r.makeFormFile(fileName, file)
	if berr != nil {
		err = ferr
		return
	}

	if err != nil {
		return "", err
	}
	done := make(chan struct{})

	ctx, cancel := context.WithTimeout(ctx, r.client.timeout)
	defer cancel()

	ticker := time.NewTicker(r.client.polling)
	defer ticker.Stop()

	go func() {
		defer close(done)

		err = retry.OnError(DefaultBackoff, isRetryable, func() error {
			localID, localErr := r.uploadFile(ctx, body, content)

			if localErr != nil {
				return errors.Wrap(localErr, "failed to get upload file req")
			}

			id = localID
			return nil
		})

		if err != nil {
			if errors.Is(err, DuplicateError) {
				err = nil
				id = ""
			}

			return
		}
	}()

	<-done
	return
}
