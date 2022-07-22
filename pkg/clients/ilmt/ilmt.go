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
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
	FetchUsageData(ctx context.Context, dateRange DateRange, selectedDate time.Time) []byte
}

func NewClient(config *IlmtConfig) Client {
	cli := &ilmtClient{
		Client:     &http.Client{},
		IlmtConfig: *config,
		req:        &reqBuilder{Host: config.Host},
	}
	return cli
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func (ilmtC *ilmtClient) FetchUsageData(ctx context.Context, dateRange DateRange, selectedDate time.Time) []byte {

	urlForStandaloneProductUsage, err := ilmtC.req.FetchUsageData(ctx, ilmtC.IlmtConfig.Host, ilmtC.IlmtConfig.Token, CRITERIA_STANDALONE, strings.Split(selectedDate.String(), BLANK_VALUE)[0], strings.Split(selectedDate.String(), BLANK_VALUE)[0])
	check(err)

	urlForProductPartOfBndlUsage, err := ilmtC.req.FetchUsageData(ctx, ilmtC.IlmtConfig.Host, ilmtC.IlmtConfig.Token, CRITEIRA_PRODUCTPARTOFBNDL, strings.Split(selectedDate.String(), BLANK_VALUE)[0], strings.Split(selectedDate.String(), BLANK_VALUE)[0])
	check(err)

	urlForParentProductUsage, err := ilmtC.req.FetchUsageData(ctx, ilmtC.IlmtConfig.Host, ilmtC.IlmtConfig.Token, CRITERIA_PARENTPRODUCT, strings.Split(selectedDate.String(), BLANK_VALUE)[0], strings.Split(selectedDate.String(), BLANK_VALUE)[0])
	check(err)

	// licence usage for product
	standaloneProductResp, err := ilmtC.Do(urlForStandaloneProductUsage)
	check(err)
	defer standaloneProductResp.Body.Close()

	standaloneProductRespData, err := ioutil.ReadAll(standaloneProductResp.Body)
	check(err)

	var standaloneProductRespObj StandaloneProductResp
	json.Unmarshal(standaloneProductRespData, &standaloneProductRespObj)

	// licence usage for product that are part of bundle
	productPartOfBndlResp, err := ilmtC.Do(urlForProductPartOfBndlUsage)
	check(err)
	defer productPartOfBndlResp.Body.Close()

	productPartOfBndlRespData, err := ioutil.ReadAll(productPartOfBndlResp.Body)
	check(err)

	var productPartOfBndlRespObj ProductPartOfBndlResp
	json.Unmarshal(productPartOfBndlRespData, &productPartOfBndlRespObj)

	// licence usage for product that are part of bundlE
	parentProductResp, err := ilmtC.Do(urlForParentProductUsage)
	check(err)
	defer parentProductResp.Body.Close()

	parentProductRespData, err := ioutil.ReadAll(parentProductResp.Body)
	check(err)

	var parentProductRespObj ParentProductResp
	json.Unmarshal(parentProductRespData, &parentProductRespObj)

	productUsageCount := len(standaloneProductRespObj.StandaloneProductLicenceUsage) + len(productPartOfBndlRespObj.ProductPartOfBndlLicenceUsage)
	productUsageTrnsfrmdEvntDataSlice := make([]ProductUsageTransformedEventData, productUsageCount)

	counter := 0
	productUsageTransformedEventJson := []byte{}

	for _, productResp := range standaloneProductRespObj.StandaloneProductLicenceUsage {

		startDateMillis := dateRange.StartDate.UnixMilli()
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

		endDateMillis := dateRange.EndDate.UnixMilli()
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
			AccountId:            "RHM account id",
			StartDate:            startDateMillis,
			EndDate:              endDateMillis,
			EventId:              eventIdFinal,
			MeasuredUsage:        measuredUsage,
			AdditionalAttributes: additionalAttributes,
		}

		productUsageTrnsfrmdEvntDataSlice[counter] = productUsageTransformedEventData
		counter++
	}

	for _, prodPartOfBundle := range productPartOfBndlRespObj.ProductPartOfBndlLicenceUsage {

		startDateMillis := dateRange.StartDate.UnixMilli()
		productId := prodPartOfBundle.ProductId
		measuredMetricId := prodPartOfBundle.MetricCodeName
		parentProductId, parentProductName, metricId, err := GetParentProduct(prodPartOfBundle, parentProductRespObj)
		if err != nil {
			log.Fatal(err.Error())
			os.Exit(1)
		}
		host := ilmtC.IlmtConfig.Host[8:38]
		BELL := '\a'

		eventId := fmt.Sprintf("%d%U%d%U%s%U%d%U%s", startDateMillis, BELL, productId, BELL, measuredMetricId, BELL, parentProductId, BELL, host)
		h := sha256.New()
		h.Write([]byte(eventId))
		bs := h.Sum(nil)
		sEnc := b64.StdEncoding.EncodeToString(bs)
		eventIdFinal := "ILMT-" + sEnc

		endDateMillis := dateRange.EndDate.UnixMilli()
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
			AccountId:            "RHM account id",
			StartDate:            startDateMillis,
			EndDate:              endDateMillis,
			EventId:              eventIdFinal,
			MeasuredUsage:        measuredUsage,
			AdditionalAttributes: additionalAttributes,
		}
		productUsageTrnsfrmdEvntDataSlice[counter] = productUsageTransformedEventData
		counter++
	}

	productUsageTransformedEvent := ProductUsageTransformedEvent{
		ProductUsageTransformedEventData: productUsageTrnsfrmdEvntDataSlice,
	}
	productUsageTransformedEventJson, _ = json.Marshal(productUsageTransformedEvent)

	return productUsageTransformedEventJson
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
