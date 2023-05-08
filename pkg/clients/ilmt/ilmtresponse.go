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
	CRITERIA_STANDALONE        string = "&limit=2&criteria={'and':[['bundle_id','<=','0']]}&criteria={'and':[['bundle_type','=','-1']]}"
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

type StandaloneProductResp struct {
	TotalRows                     int                             `json:"total"`
	StandaloneProductLicenceUsage []StandaloneProductLicenceUsage `json:"rows"`
}

type StandaloneProductLicenceUsage struct {
	ProductId            int64  `json:"product_id"`
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

type ProductPartOfBndlResp struct {
	TotalRows                     int                             `json:"total"`
	ProductPartOfBndlLicenceUsage []ProductPartOfBndlLicenceUsage `json:"rows"`
}

type ProductPartOfBndlLicenceUsage struct {
	ProductId            int64  `json:"product_id"`
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

type ProductUsageTransformedEvent struct {
	ProductUsageTransformedEventData []ProductUsageTransformedEventData `json:"data"`
}

type ProductUsageTransformedEventData struct {
	//AccountId            string `json:"accountId"`
	AdditionalAttributes `json:"additionalAttributes"`
	EndDate              int64           `json:"end"`
	StartDate            int64           `json:"start"`
	EventId              string          `json:"eventId"`
	MeasuredUsage        []MeasuredUsage `json:"measuredUsage"`
}

type MeasuredUsage struct {
	MetricId string  `json:"metricId"`
	Value    float64 `json:"value"`
}

type AdditionalAttributes struct {
	HostName               string `json:"hostname"`
	MeasuredMetricId       string `json:"measuredMetricId"`
	MeasuredValue          int    `json:"measuredValue"`
	MetricType             string `json:"metricType"`
	ParentProductId        int64  `json:"parentProductId,omitempty"`
	ParentProductName      string `json:"parentProductName,omitempty"`
	ProductConversionRatio string `json:"productConversionRatio,omitempty"`
	ProductId              int64  `json:"productId"`
	ProductName            string `json:"productName"`
	Source                 string `json:"source"`
}

type ParentProductResp struct {
	TotalRows                 int                         `json:"total"`
	ParentProductLicenceUsage []ParentProductLicenceUsage `json:"rows"`
}

type ParentProductLicenceUsage struct {
	ProductId            int64  `json:"product_id"`
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
