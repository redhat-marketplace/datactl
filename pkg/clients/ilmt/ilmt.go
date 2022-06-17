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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-logr/logr"
	"k8s.io/klog/v2/klogr"
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
	urlForStandaloneProductUsage, err := ilmtC.req.FetchUsageData(ctx, ilmtC.IlmtConfig.Host, ilmtC.IlmtConfig.Token, CRITERIA_STANDALONE, dateRange.StartDate, dateRange.EndDate)
	if err != nil {
		log.Fatal(err.Error())
		return -1, EMPTY, err
	}

	urlForProductPartOfBndlUsage, err := ilmtC.req.FetchUsageData(ctx, ilmtC.IlmtConfig.Host, ilmtC.IlmtConfig.Token, CRITEIRA_PRODUCTPARTOFBNDL, dateRange.StartDate, dateRange.EndDate)
	if err != nil {
		log.Fatal(err.Error())
		return -1, EMPTY, err
	}

	urlForParentProductUsage, err := ilmtC.req.FetchUsageData(ctx, ilmtC.IlmtConfig.Host, ilmtC.IlmtConfig.Token, CRITERIA_PARENTPRODUCT, dateRange.StartDate, dateRange.EndDate)
	if err != nil {
		log.Fatal(err.Error())
		return -1, EMPTY, err
	}

	// licence usage for product
	productResponse, err := ilmtC.Do(urlForStandaloneProductUsage)

	if err != nil {
		log.Fatal(err.Error())
		return -1, EMPTY, err
	}

	productResponseData, err := ioutil.ReadAll(productResponse.Body)
	if err != nil {
		log.Fatal(err.Error())
		return -1, EMPTY, err
	}
	// fmt.Println(string(productResponseData))

	var responseObject ProductResponse
	json.Unmarshal(productResponseData, &responseObject)

	fmt.Println("")
	fmt.Println("Count of total no of standalone products for usage: ", responseObject.TotalRows)
	fmt.Println("Count of standalone products usage fetched: ", len(responseObject.ProductLicenceUsage))

	fmt.Println("Standalone Product Name List: ")
	for i := 0; i < len(responseObject.ProductLicenceUsage); i++ {
		fmt.Println(responseObject.ProductLicenceUsage[i].ProductName)
	}

	// licence usage for product that are part of bundle
	prodPartOfBndlResp, err := ilmtC.Do(urlForProductPartOfBndlUsage)
	if err != nil {
		log.Fatal(err.Error())
		return -1, EMPTY, err
	}

	prodPartOfBndlRespData, err := ioutil.ReadAll(prodPartOfBndlResp.Body)
	if err != nil {
		log.Fatal(err.Error())
		return -1, EMPTY, err
	}
	// fmt.Println(string(prodPartOfBndlRespData))

	var respObjPartOfBndl ProdPartOfBndlResp
	json.Unmarshal(prodPartOfBndlRespData, &respObjPartOfBndl)

	fmt.Println("")
	fmt.Println("===============================================================================")
	fmt.Println("Count of total no of products part of bundle for usage: ", respObjPartOfBndl.TotalRows)
	fmt.Println("Count of products part of bundle usage fetched: ", len(respObjPartOfBndl.ProdPartOfBundlLicenceUsage))

	fmt.Println("Product Part Of bundle Name List: ")
	for i := 0; i < len(respObjPartOfBndl.ProdPartOfBundlLicenceUsage); i++ {
		fmt.Println(respObjPartOfBndl.ProdPartOfBundlLicenceUsage[i].ProductName)
	}

	// licence usage for product that are part of bundlE
	parentProdResp, err := ilmtC.Do(urlForParentProductUsage)

	if err != nil {
		log.Fatal(err.Error())
		return -1, EMPTY, err
	}

	parentProdRespData, err := ioutil.ReadAll(parentProdResp.Body)
	if err != nil {
		log.Fatal(err.Error())
		return -1, EMPTY, err
	}
	// fmt.Println(string(parentProdRespData))

	var respObjParent ParentProdResp
	json.Unmarshal(parentProdRespData, &respObjParent)

	fmt.Println("")
	fmt.Println("===============================================================================")
	fmt.Println("Count of total no of parent products for usage: ", respObjParent.TotalRows)
	fmt.Println("Count of parent products usage fetched: ", len(respObjParent.ParentProdLicenceUsage))

	fmt.Println("Parent Product Name List: ")
	for i := 0; i < len(respObjParent.ParentProdLicenceUsage); i++ {
		fmt.Println(respObjParent.ParentProdLicenceUsage[i].ProductName)
	}
	fmt.Println("")

	productUsageRespStr := string(productResponseData) + string(prodPartOfBndlRespData) + string(parentProdRespData)
	productCount := respObjParent.TotalRows + respObjPartOfBndl.TotalRows + responseObject.TotalRows
	return productCount, productUsageRespStr, nil
}

type reqBuilder struct {
	Host string
}

func (u *reqBuilder) FetchUsageData(ctx context.Context, host string, token string, criteria string, startdate string, enddate string) (*http.Request, error) {
	u.Host = fmt.Sprintf("%s/api/sam/v2/license_usage?token=%s%s%s%s%s%s%s", host, token, COLUMN_NAMES_APPEND, criteria, START_DATE_FLD_APPEND, startdate, END_DATE_FLD_APPEND, enddate)
	return http.NewRequestWithContext(ctx, http.MethodGet, u.Host, nil)
}
