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
	Tar(tempDir, &buffer)

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

// Tar takes a source and variable writers and walks 'source' writing each file
// found to the tar writer; the purpose for accepting multiple writers is to allow
// for multiple outputs (for example a file, or md5 hash)
func Tar(src string, writers ...io.Writer) error {
	// ensure the src actually exists before trying to tar it
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("unable to tar files - %v", err.Error())
	}

	mw := io.MultiWriter(writers...)

	gzw := gzip.NewWriter(mw)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// walk path
	return filepath.Walk(src, func(file string, fi os.FileInfo, errIn error) error {

		fmt.Println(file)

		// return on any error
		if errIn != nil {
			return errors.Wrap(errIn, "fail to tar files")
		}

		// return on non-regular files (thanks to [kumo](https://medium.com/@komuw/just-like-you-did-fbdd7df829d3) for this suggested update)
		if !fi.Mode().IsRegular() {
			return nil
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return errors.Wrap(err, "failed to create new dir")
		}

		// update the name to correctly reflect the desired destination when untaring
		header.Name = strings.TrimPrefix(strings.ReplaceAll(file, src, ""), string(filepath.Separator))

		// write the header
		if err = tw.WriteHeader(header); err != nil {
			return errors.Wrap(err, "failed to write header")
		}

		// open files for taring
		f, err := os.Open(file)
		if err != nil {
			return errors.Wrap(err, "fail to open file for taring")
		}

		// copy file data into tar writer
		if _, err := io.Copy(tw, f); err != nil {
			return errors.Wrap(err, "failed to copy data")
		}

		// manually close here after each file operation; defering would cause each file close
		// to wait until all operations have completed.
		f.Close()

		return nil
	})
}
