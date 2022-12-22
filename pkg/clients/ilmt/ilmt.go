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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
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
		Client:     &http.Client{},
		IlmtConfig: *config,
		req:        &reqBuilder{Host: config.Host},
	}
	return cli
}

func (ilmtC *ilmtClient) FetchUsageData(ctx context.Context, dateRange DateRange) (int, string, error) {
	startDate, _ := time.Parse(REQUIRED_FORMAT, dateRange.StartDate)
	endDate, _ := time.Parse(REQUIRED_FORMAT, dateRange.EndDate)
	fileCounter := 0

	productUsageTransformedEventJsonNew := make([]ProductUsageTransformedEventData, 0)

	for selectedDate := startDate; !selectedDate.After(endDate); selectedDate = selectedDate.AddDate(0, 0, 1) {

		urlForStandaloneProductUsage, err := ilmtC.req.FetchUsageData(ctx, ilmtC.IlmtConfig.Host, ilmtC.IlmtConfig.Token, CRITERIA_STANDALONE, strings.Split(selectedDate.String(), BLANK_VALUE)[0], strings.Split(selectedDate.String(), BLANK_VALUE)[0])
		if err != nil {
			log.Fatal(err.Error())
			return -1, EMPTY, err
		}

		urlForProductPartOfBndlUsage, err := ilmtC.req.FetchUsageData(ctx, ilmtC.IlmtConfig.Host, ilmtC.IlmtConfig.Token, CRITEIRA_PRODUCTPARTOFBNDL, strings.Split(selectedDate.String(), BLANK_VALUE)[0], strings.Split(selectedDate.String(), BLANK_VALUE)[0])
		if err != nil {
			log.Fatal(err.Error())
			return -1, EMPTY, err
		}

		urlForParentProductUsage, err := ilmtC.req.FetchUsageData(ctx, ilmtC.IlmtConfig.Host, ilmtC.IlmtConfig.Token, CRITERIA_PARENTPRODUCT, strings.Split(selectedDate.String(), BLANK_VALUE)[0], strings.Split(selectedDate.String(), BLANK_VALUE)[0])
		if err != nil {
			log.Fatal(err.Error())
			return -1, EMPTY, err
		}

		// licence usage for product
		standaloneProductResp, err := ilmtC.Do(urlForStandaloneProductUsage)

		if err != nil {
			log.Fatal(err.Error())
			return -1, EMPTY, err
		}

		standaloneProductRespData, err := ioutil.ReadAll(standaloneProductResp.Body)
		if err != nil {
			log.Fatal(err.Error())
			return -1, EMPTY, err
		}

		var standaloneProductRespObj StandaloneProductResp
		json.Unmarshal(standaloneProductRespData, &standaloneProductRespObj)

		// licence usage for product that are part of bundle
		productPartOfBndlResp, err := ilmtC.Do(urlForProductPartOfBndlUsage)
		if err != nil {
			log.Fatal(err.Error())
			return -1, EMPTY, err
		}

		productPartOfBndlRespData, err := ioutil.ReadAll(productPartOfBndlResp.Body)
		if err != nil {
			log.Fatal(err.Error())
			return -1, EMPTY, err
		}

		var productPartOfBndlRespObj ProductPartOfBndlResp
		json.Unmarshal(productPartOfBndlRespData, &productPartOfBndlRespObj)

		// licence usage for product that are part of bundlE
		parentProductResp, err := ilmtC.Do(urlForParentProductUsage)
		if err != nil {
			log.Fatal(err.Error())
			return -1, EMPTY, err
		}

		parentProductRespData, err := ioutil.ReadAll(parentProductResp.Body)
		if err != nil {
			log.Fatal(err.Error())
			return -1, EMPTY, err
		}

		var parentProductRespObj ParentProductResp
		json.Unmarshal(parentProductRespData, &parentProductRespObj)

		for _, productResp := range standaloneProductRespObj.StandaloneProductLicenceUsage {

			startDateMillis := startDate.UnixMilli()
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

			endDateMillis := endDate.UnixMilli()
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

			productUsageTransformedEventJsonNew = append(productUsageTransformedEventJsonNew, productUsageTransformedEventData)
		}

		for _, prodPartOfBundle := range productPartOfBndlRespObj.ProductPartOfBndlLicenceUsage {

			startDateMillis := startDate.UnixMilli()
			endDateMillis := endDate.UnixMilli()
			productId := prodPartOfBundle.ProductId
			measuredMetricId := prodPartOfBundle.MetricCodeName
			parentProductId, parentProductName, metricId, err := GetParentProduct(prodPartOfBundle, parentProductRespObj)
			if err != nil {
				log.Fatal(err.Error())
				os.Exit(1)
			}
			host := ilmtC.IlmtConfig.Host[8:38]
			BELL := '\a'

			eventId := fmt.Sprintf("%d-%d%U%d%U%s%U%d%U%s", startDateMillis, endDateMillis, BELL, productId, BELL, measuredMetricId, BELL, parentProductId, BELL, host)
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
			productUsageTransformedEventJsonNew = append(productUsageTransformedEventJsonNew, productUsageTransformedEventData)
		}

		fileCounter++
	}

	productUsageTransformedEventNew := ProductUsageTransformedEvent{
		ProductUsageTransformedEventData: productUsageTransformedEventJsonNew,
	}
	productUsageTransformedEventNewJson, _ := json.Marshal(productUsageTransformedEventNew)
	productUsageTransformedEventNewJsonStr := string(productUsageTransformedEventNewJson)

	// fix to adjust types returned by ILMT and required by RHM
	measuredValue := regexp.MustCompile(`"measuredValue":\s?(\d*),`)
	parentProductId := regexp.MustCompile(`"parentProductId":\s?(\d*),`)
	productId := regexp.MustCompile(`"productId":\s?(\d*),`)

	productUsageTransformedEventNewJsonStr = measuredValue.ReplaceAllString(productUsageTransformedEventNewJsonStr, "\"measuredValue\":\"$1\",")
	productUsageTransformedEventNewJsonStr = parentProductId.ReplaceAllString(productUsageTransformedEventNewJsonStr, "\"parentProductId\":\"$1\",")
	productUsageTransformedEventNewJsonStr = productId.ReplaceAllString(productUsageTransformedEventNewJsonStr, "\"productId\":\"$1\",")

	return fileCounter, productUsageTransformedEventNewJsonStr, nil
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
