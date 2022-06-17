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

const (
	COLUMN_NAMES_APPEND        string = "&columns[]=product_id&columns[]=product_name&columns[]=metric_code_name&columns[]=bundle_id&columns[]=flex_id&columns[]=bundle_type&columns[]=bundle_name&columns[]=hwm_quantity&columns[]=bundle_metric_contribution&columns[]=product_bundle_ratio_factor&columns[]=product_bundle_ratio_divider"
	CRITERIA_STANDALONE        string = "&criteria={'and':[['bundle_id','<=','0']]}&criteria={'and':[['bundle_type','=','-1']]}"
	CRITEIRA_PRODUCTPARTOFBNDL string = "&criteria={'and':[['bundle_id','>','0']]}"
	CRITERIA_PARENTPRODUCT     string = "&criteria={'and':[['bundle_type','>','-1']]}"
	START_DATE_FLD_APPEND      string = "&startdate="
	END_DATE_FLD_APPEND        string = "&enddate="
	EMPTY                      string = ""
)

type DateRange struct {
	StartDate string
	EndDate   string
}

type ProductLicenceUsage struct {
	ProductId            int    `json:"product_id"`
	ProductName          string `json:"product_name"`
	MetricCodeName       string `json:"metric_code_name"`
	HwmQuantity          int    `json:"hwm_quantity"`
	BundleId             int    `json:"bundle_id"`
	FlexId               int    `json:"flex_id"`
	BundleType           int    `json:"bundle_type"`
	BundleName           string `json:"bundle_name"`
	BundleMetricContrbtn int    `json:"bundle_metric_contribution"`
	ProdBndlRatioFactor  int    `json:"product_bundle_ratio_factor"`
	ProdBndlRatioDivider int    `json:"product_bundle_ratio_divider"`
}

type ProductResponse struct {
	TotalRows           int                   `json:"total"`
	ProductLicenceUsage []ProductLicenceUsage `json:"rows"`
}

type ProdPartOfBndlResp struct {
	TotalRows                   int                           `json:"total"`
	ProdPartOfBundlLicenceUsage []ProdPartOfBundlLicenceUsage `json:"rows"`
}

type ProdPartOfBundlLicenceUsage struct {
	ProductId            int    `json:"product_id"`
	ProductName          string `json:"product_name"`
	MetricCodeName       string `json:"metric_code_name"`
	HwmQuantity          int    `json:"hwm_quantity"`
	BundleId             int    `json:"bundle_id"`
	FlexId               int    `json:"flex_id"`
	BundleType           int    `json:"bundle_type"`
	BundleName           string `json:"bundle_name"`
	BundleMetricContrbtn int    `json:"bundle_metric_contribution"`
	ProdBndlRatioFactor  int    `json:"product_bundle_ratio_factor"`
	ProdBndlRatioDivider int    `json:"product_bundle_ratio_divider"`
}

type ParentProdResp struct {
	TotalRows              int                      `json:"total"`
	ParentProdLicenceUsage []ParentProdLicenceUsage `json:"rows"`
}

type ParentProdLicenceUsage struct {
	ProductId            int    `json:"product_id"`
	ProductName          string `json:"product_name"`
	MetricCodeName       string `json:"metric_code_name"`
	HwmQuantity          int    `json:"hwm_quantity"`
	BundleId             int    `json:"bundle_id"`
	FlexId               int    `json:"flex_id"`
	BundleType           int    `json:"bundle_type"`
	BundleName           string `json:"bundle_name"`
	BundleMetricContrbtn int    `json:"bundle_metric_contribution"`
	ProdBndlRatioFactor  int    `json:"product_bundle_ratio_factor"`
	ProdBndlRatioDivider int    `json:"product_bundle_ratio_divider"`
}
