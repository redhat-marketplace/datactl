package sources

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"emperror.dev/errors"
	"github.com/redhat-marketplace/datactl/pkg/bundle"
	"github.com/redhat-marketplace/datactl/pkg/clients/ilmt"
	"github.com/redhat-marketplace/datactl/pkg/datactl/api"
	dataservicev1 "github.com/redhat-marketplace/datactl/pkg/datactl/api/dataservice/v1"
	"github.com/redhat-marketplace/datactl/pkg/printers"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	_, productUsageRespStr, err := i.ilmt.FetchUsageData(ctx, dateRangeOptions)

	if err != nil {
		return -1, err
	}

	i.productUsageResponseStr = productUsageRespStr

	// create temporary directory
	tempDir, err := ioutil.TempDir("", "iltmdata")
	if err != nil {
		return 0, err
	}

	// create file with received data and manifest in temporary directory
	err = CreateFileFromString(filepath.Join(tempDir, "ilmtdata.json"), productUsageRespStr)
	if err != nil {
		return 0, err
	}

	CreateFileFromString(filepath.Join(tempDir, "manifest.json"), "{\"version\":\"1\",\"type\":\"accountMetrics\"}")
	if err != nil {
		return 0, err
	}

	var buffer bytes.Buffer

	// create archive file
	err = Tar(tempDir, &buffer)
	if err != nil {
		return 0, err
	}

	reportFileName := fmt.Sprintf("upload-ilmt-%s-%s.tar.gz", dateRangeOptions.StartDate, dateRangeOptions.EndDate)

	// create new bundle file
	w, err := bundle.NewFile(reportFileName, int64(buffer.Len()))
	if err != nil {
		return 0, err
	}

	w.Write(buffer.Bytes())

	// remove temporary directory
	defer os.RemoveAll(tempDir)

	ilmtFile := &dataservicev1.FileInfoCTLAction{
		Action: dataservicev1.Pull,
		FileInfo: &dataservicev1.FileInfo{
			Source:     "ilmt",
			SourceType: "report",
			Size:       uint32(buffer.Len()),
			MimeType:   "application/gzip",
			CreatedAt:  &v1.Time{},
		},
	}
	ilmtFile.Name = reportFileName

	currentMeteringExport.Files = append(currentMeteringExport.Files, ilmtFile)

	return 1, nil
}

// Creates file with given data content
func CreateFileFromString(outFile string, data string) error {
	f, err := os.Create(outFile)
	if err != nil {
		return err
	}

	_, err = f.WriteString(data)

	if err != nil {
		f.Close()
		return err
	}

	f.Close()

	return nil
}

func Tar(src string, writers ...io.Writer) error {
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("unable to tar files - %v", err.Error())
	}

	mw := io.MultiWriter(writers...)

	gzw := gzip.NewWriter(mw)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	return filepath.Walk(src, func(file string, fi os.FileInfo, errIn error) error {

		if errIn != nil {
			return errors.Wrap(errIn, "failed to tar files")
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return errors.Wrap(err, "failed to create new dir")
		}

		header.Name = strings.TrimPrefix(strings.ReplaceAll(file, src, ""), string(filepath.Separator))

		if err = tw.WriteHeader(header); err != nil {
			return errors.Wrap(err, "failed to write header")
		}

		f, err := os.Open(file)
		if err != nil {
			return errors.Wrap(err, "failed to open file for taring")
		}

		if _, err := io.Copy(tw, f); err != nil {
			return errors.Wrap(err, "failed to copy data")
		}

		f.Close()

		return nil
	})
}
