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

package ilmt

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/redhat-marketplace/datactl/pkg/clients/shared"
	"k8s.io/klog/v2/klogr"
)

const (
	REQUIRED_FORMAT string = "2006-01-02"
	BLANK_VALUE     string = " "
)

var (
	logger logr.Logger = klogr.New().V(5).WithName("pkg/clients/ilmt")
)

type IlmtConfig struct {
	Host  string `json:"host"`
	Port  string `json:"port"`
	Token string `json:"token"`

	TlsConfig *tls.Config
}

type ilmtClient struct {
	*http.Client
	IlmtConfig
	req *reqBuilder
}

type Client interface {
	FetchUsageData(ctx context.Context, dateRange DateRange) (int, string, error)
}

func NewClient(config *IlmtConfig) Client {
	cli := &ilmtClient{
		Client: shared.NewHttpClient(
			config.TlsConfig,
		),
		IlmtConfig: *config,
		req:        &reqBuilder{Host: config.Host},
	}
	return cli
}

func (ilmtC *ilmtClient) FetchUsageData(ctx context.Context, dateRange DateRange) (int, string, error) {
	startDate, _ := time.Parse(REQUIRED_FORMAT, dateRange.StartDate)
	endDate, _ := time.Parse(REQUIRED_FORMAT, dateRange.EndDate)
	fileCounter := 0

	productUsageTransformedEventJson := make([]ProductUsageTransformedEventData, 0)

	for selectedDate := startDate; !selectedDate.After(endDate); selectedDate = selectedDate.AddDate(0, 0, 1) {

		startDateMillis := selectedDate.UnixMilli()
		endDateMillis := selectedDate.AddDate(0, 0, 1).UnixMilli() - 1

		urlForStandaloneProductUsage, err := ilmtC.req.FetchUsageData(ctx, ilmtC.IlmtConfig.Host, ilmtC.IlmtConfig.Token, CRITERIA_STANDALONE, strings.Split(selectedDate.String(), BLANK_VALUE)[0], strings.Split(selectedDate.String(), BLANK_VALUE)[0])
		if err != nil {
			log.Fatal(err.Error())
		}

		urlForProductPartOfBndlUsage, err := ilmtC.req.FetchUsageData(ctx, ilmtC.IlmtConfig.Host, ilmtC.IlmtConfig.Token, CRITEIRA_PRODUCTPARTOFBNDL, strings.Split(selectedDate.String(), BLANK_VALUE)[0], strings.Split(selectedDate.String(), BLANK_VALUE)[0])
		if err != nil {
			log.Fatal(err.Error())
		}

		urlForParentProductUsage, err := ilmtC.req.FetchUsageData(ctx, ilmtC.IlmtConfig.Host, ilmtC.IlmtConfig.Token, CRITERIA_PARENTPRODUCT, strings.Split(selectedDate.String(), BLANK_VALUE)[0], strings.Split(selectedDate.String(), BLANK_VALUE)[0])
		if err != nil {
			log.Fatal(err.Error())
		}

		// licence usage for product
		standaloneProductResp, err := ilmtC.Do(urlForStandaloneProductUsage)

		if err != nil {
			log.Fatal(err.Error())
		}

		standaloneProductRespData, err := ioutil.ReadAll(standaloneProductResp.Body)
		if err != nil {
			log.Fatal(err.Error())
		}

		var standaloneProductRespObj StandaloneProductResp
		json.Unmarshal(standaloneProductRespData, &standaloneProductRespObj)

		// licence usage for product that are part of bundle
		productPartOfBndlResp, err := ilmtC.Do(urlForProductPartOfBndlUsage)
		if err != nil {
			log.Fatal(err.Error())
		}

		productPartOfBndlRespData, err := ioutil.ReadAll(productPartOfBndlResp.Body)
		if err != nil {
			log.Fatal(err.Error())
		}

		var productPartOfBndlRespObj ProductPartOfBndlResp
		json.Unmarshal(productPartOfBndlRespData, &productPartOfBndlRespObj)

		// licence usage for product that are part of bundlE
		parentProductResp, err := ilmtC.Do(urlForParentProductUsage)
		if err != nil {
			log.Fatal(err.Error())
		}

		parentProductRespData, err := ioutil.ReadAll(parentProductResp.Body)
		if err != nil {
			log.Fatal(err.Error())
		}

		var parentProductRespObj ParentProductResp
		json.Unmarshal(parentProductRespData, &parentProductRespObj)

		for _, productResp := range standaloneProductRespObj.StandaloneProductLicenceUsage {

			productId := productResp.ProductId
			measuredMetricId := productResp.MetricCodeName
			metricId := productResp.MetricCodeName
			host := ilmtC.IlmtConfig.Host[8:38]
			BELL := '\a'

			eventId := fmt.Sprintf("%d%U%d%U%s%U%s%U%s", startDateMillis, BELL, productId, BELL, measuredMetricId, BELL, EMPTY, BELL, host)
			h := sha256.New()
			h.Write([]byte(eventId))
			bs := h.Sum(nil)
			sEnc := b64.StdEncoding.EncodeToString(bs)
			eventIdFinal := "ILMT-" + sEnc

			measuredValue := productResp.HwmQuantity
			productName := productResp.ProductName

			measuredUsage := []MeasuredUsage{
				{
					MetricId: metricId,
					Value:    1,
				},
			}

			additionalAttributes := AdditionalAttributes{
				HostName:         host,
				MeasuredMetricId: measuredMetricId,
				MeasuredValue:    measuredValue,
				MetricType:       "license",
				ProductId:        productId,
				ProductName:      productName,
				Source:           "ILMT",
			}

			productUsageTransformedEventData := ProductUsageTransformedEventData{
				StartDate:            startDateMillis,
				EndDate:              endDateMillis,
				EventId:              eventIdFinal,
				MeasuredUsage:        measuredUsage,
				AdditionalAttributes: additionalAttributes,
			}

			productUsageTransformedEventJson = append(productUsageTransformedEventJson, productUsageTransformedEventData)
		}

		for _, prodPartOfBundle := range productPartOfBndlRespObj.ProductPartOfBndlLicenceUsage {

			productId := prodPartOfBundle.ProductId
			measuredMetricId := prodPartOfBundle.MetricCodeName
			parentProductId, parentProductName, metricId, err := GetParentProduct(prodPartOfBundle, parentProductRespObj)
			if err != nil {
				log.Fatal(err.Error())
			}
			host := ilmtC.IlmtConfig.Host[8:38]
			BELL := '\a'

			eventId := fmt.Sprintf("%d%U%d%U%s%U%d%U%s", startDateMillis, BELL, productId, BELL, measuredMetricId, BELL, parentProductId, BELL, host)
			h := sha256.New()
			h.Write([]byte(eventId))
			bs := h.Sum(nil)
			sEnc := b64.StdEncoding.EncodeToString(bs)
			eventIdFinal := "ILMT-" + sEnc

			measuredValue := prodPartOfBundle.HwmQuantity
			productConversionRatio := GetProductConversionRatio(prodPartOfBundle.ProdBndlRatioDivider, prodPartOfBundle.ProdBndlRatioFactor)
			productName := prodPartOfBundle.ProductName

			measuredUsage := []MeasuredUsage{
				{
					MetricId: metricId,
					Value:    1,
				},
			}

			additionalAttributes := AdditionalAttributes{
				HostName:               host,
				MeasuredMetricId:       measuredMetricId,
				MeasuredValue:          measuredValue,
				MetricType:             "license",
				ParentProductId:        parentProductId,
				ParentProductName:      parentProductName,
				ProductConversionRatio: productConversionRatio,
				ProductId:              productId,
				ProductName:            productName,
				Source:                 "ILMT",
			}

			productUsageTransformedEventData := ProductUsageTransformedEventData{
				StartDate:            startDateMillis,
				EndDate:              endDateMillis,
				EventId:              eventIdFinal,
				MeasuredUsage:        measuredUsage,
				AdditionalAttributes: additionalAttributes,
			}
			productUsageTransformedEventJson = append(productUsageTransformedEventJson, productUsageTransformedEventData)
		}

		fileCounter++
	}

	productUsageTransformedEvent := ProductUsageTransformedEvent{
		ProductUsageTransformedEventData: productUsageTransformedEventJson,
	}
	productUsageTransformedEventNewJson, _ := json.Marshal(productUsageTransformedEvent)
	productUsageTransformedEventJsonStr := string(productUsageTransformedEventNewJson)

	// fix to adjust types returned by ILMT and required by RHM
	measuredValue := regexp.MustCompile(`"measuredValue":\s?(\d*),`)
	parentProductId := regexp.MustCompile(`"parentProductId":\s?(\d*),`)
	productId := regexp.MustCompile(`"productId":\s?(\d*),`)

	productUsageTransformedEventJsonStr = measuredValue.ReplaceAllString(productUsageTransformedEventJsonStr, "\"measuredValue\":\"$1\",")
	productUsageTransformedEventJsonStr = parentProductId.ReplaceAllString(productUsageTransformedEventJsonStr, "\"parentProductId\":\"$1\",")
	productUsageTransformedEventJsonStr = productId.ReplaceAllString(productUsageTransformedEventJsonStr, "\"productId\":\"$1\",")

	return fileCounter, productUsageTransformedEventJsonStr, nil
}

func GetParentProduct(prodPartOfbndl ProductPartOfBndlLicenceUsage, parentProdResp ParentProductResp) (int64, string, string, error) {
	for _, parentProdLicUsage := range parentProdResp.ParentProductLicenceUsage {
		if prodPartOfbndl.BundleId == parentProdLicUsage.FlexId {
			return parentProdLicUsage.ProductId, parentProdLicUsage.ProductName, parentProdLicUsage.MetricCodeName, nil
		}
	}
	parentProdError := fmt.Sprintf("Parent product not found of child product with id %d and product name with %s", prodPartOfbndl.ProductId, prodPartOfbndl.ProductName)
	return -1, EMPTY, EMPTY, errors.New(parentProdError)
}

func GetProductConversionRatio(prodBndlRatioDivider int, prodBndlRatioFactor int) string {
	if prodBndlRatioDivider != 0 && prodBndlRatioFactor != 0 {
		prodBndlRatioDivider := strconv.Itoa(prodBndlRatioDivider)
		prodBndlRatioFactor := strconv.Itoa(prodBndlRatioFactor)
		return (prodBndlRatioDivider + ":" + prodBndlRatioFactor)
	} else {
		return EMPTY
	}
}

type reqBuilder struct {
	Host string
}

func (u *reqBuilder) FetchUsageData(ctx context.Context, host string, token string, criteria string, startdate string, enddate string) (*http.Request, error) {
	u.Host = fmt.Sprintf("%s/api/sam/v2/license_usage?token=%s%s%s%s%s%s%s", host, token, COLUMN_NAMES_APPEND, criteria, START_DATE_FLD_APPEND, startdate, END_DATE_FLD_APPEND, enddate)
	return http.NewRequestWithContext(ctx, http.MethodGet, u.Host, nil)
}

func Date(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
