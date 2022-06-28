package sources

import (
	"context"

	"github.com/redhat-marketplace/datactl/pkg/bundle"
	"github.com/redhat-marketplace/datactl/pkg/clients/ilmt"
	"github.com/redhat-marketplace/datactl/pkg/datactl/api"
	"github.com/redhat-marketplace/datactl/pkg/printers"
)

const (
	StartDate = "startDate"
	EndDate   = "endDate"
	EMPTY     = ""
)

type ilmtSource struct {
	printers.TablePrinter
	ilmt                    ilmt.Client
	productUsageResponseStr string
}

func NewIlmtSource(
	ilmt ilmt.Client,
	printer printers.TablePrinter,
) (Source, error) {
	i := &ilmtSource{
		ilmt:         ilmt,
		TablePrinter: printer,
	}
	return i, nil
}

func (i *ilmtSource) GetResponse() string {
	return i.productUsageResponseStr
}

func (i *ilmtSource) Pull(
	ctx context.Context,
	currentMeteringExport *api.MeteringExport,
	bundle *bundle.BundleFile,
	options GenericOptions,
) (int, error) {
	startDate, _, err := options.GetString(StartDate)
	if err != nil {
		return 0, err
	}

	endDate, _, err := options.GetString(EndDate)
	if err != nil {
		return 0, err
	}

	dateRangeOptions := ilmt.DateRange{
		StartDate: startDate,
		EndDate:   endDate,
	}

	productCount, productUsageRespStr, err := i.ilmt.FetchUsageData(ctx, dateRangeOptions)

	if err != nil {
		return -1, err
	}

	i.productUsageResponseStr = productUsageRespStr
	return productCount, nil
}
