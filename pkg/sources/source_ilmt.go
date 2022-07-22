package sources

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/redhat-marketplace/datactl/pkg/bundle"
	"github.com/redhat-marketplace/datactl/pkg/clients/ilmt"
	"github.com/redhat-marketplace/datactl/pkg/datactl/api"
	ilmtv1 "github.com/redhat-marketplace/datactl/pkg/datactl/api/dataservice/v1"
	"github.com/redhat-marketplace/datactl/pkg/printers"
)

const (
	StartDate              = "startDate"
	EndDate                = "endDate"
	EMPTY                  = ""
	REQUIRED_FORMAT        = "2006-01-02"
	SourceC                = "source"
	SourceType             = "sourceType"
	SPACE                  = " "
	FORMAT                 = "20060102T150405Z"
	BUNDLETEMPDIRNAME      = "rhm-upload-%s"
	HOME                   = "HOME"
	TEMPDIRPART            = "/go/src/datactl/pkg"
	ARCHIVEFILETEMPDIRNAME = "upload-%s"
	RESPONSEFILE           = "%s/%s.json"
	MANIFESTFILE           = "%s/manifest.json"
	MANIFESTFILECONTENT    = `{"version":"1","type":"accountMetrics"}`
	BUNDLEDIRPART          = "/.datactl/data"
	TARGZ                  = ".tar.gz"
	TARGETGZ               = "%s.tar.gz"
	TARGET                 = "%s.tar"
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

func check(e error) {
	if e != nil {
		panic(e)
	}
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

	source, _, err := options.GetString(SourceC)
	if err != nil {
		return 0, err
	}

	sourceType, _, err := options.GetString(SourceType)
	if err != nil {
		return 0, err
	}

	startDateT, _ := time.Parse(REQUIRED_FORMAT, startDate)
	endDateT, _ := time.Parse(REQUIRED_FORMAT, endDate)

	dateRangeOptions := ilmt.DateRange{
		StartDate: startDateT,
		EndDate:   endDateT,
	}

	fileCounter := 0
	files := []*ilmtv1.FileInfoCTLAction{}
	found := 0
	pulled := 0

	timestamp := time.Now().Format(FORMAT)
	bundleTempDirName := fmt.Sprintf(BUNDLETEMPDIRNAME, timestamp)
	tempDir := os.Getenv(HOME)
	tempDirPart := TEMPDIRPART
	tempDir += tempDirPart
	bundleTempDir, err := os.MkdirTemp(tempDir, bundleTempDirName)
	check(err)
	defer os.RemoveAll(bundleTempDir)

	for selectedDate := startDateT; !selectedDate.After(endDateT); selectedDate = selectedDate.AddDate(0, 0, 1) {
		selDate := strings.Split(selectedDate.String(), SPACE)[0]
		archFileTempDirName := fmt.Sprintf(ARCHIVEFILETEMPDIRNAME, selDate)
		archFileTempDir, err := os.MkdirTemp(tempDir, archFileTempDirName)
		check(err)
		defer os.RemoveAll(archFileTempDir)

		info := &ilmtv1.FileInfo{
			Id:         strconv.Itoa(fileCounter),
			Source:     source,
			SourceType: sourceType,
		}

		cliFile := ilmtv1.NewFileInfoCTLAction(info)

		response := i.ilmt.FetchUsageData(ctx, dateRangeOptions, selectedDate)
		check(err)

		responseFile := fmt.Sprintf(RESPONSEFILE, archFileTempDir, selDate)
		manifestFile := fmt.Sprintf(MANIFESTFILE, archFileTempDir)
		manifestFileContent := MANIFESTFILECONTENT
		err = ioutil.WriteFile(responseFile, response, 0644)
		check(err)
		err = ioutil.WriteFile(manifestFile, []byte(manifestFileContent), 0644)
		check(err)

		makeArchive(archFileTempDir, bundleTempDir)

		cliFile.Name = filepath.Base(archFileTempDir) + TARGZ
		archiveFileSize, err := os.Stat(bundleTempDir)
		check(err)
		cliFile.Size = uint32(archiveFileSize.Size())
		cliFile.Action = ilmtv1.Pull
		cliFile.Result = ilmtv1.Ok

		files = append(files, cliFile)
		found = found + 1
		pulled = pulled + 1

		i.TableOutput(func(tosp printers.PrintObj) {
			tosp.Print(cliFile)
		})
		fileCounter++
	}

	bundleDir := os.Getenv(HOME)
	bundlepDirPart := BUNDLEDIRPART
	bundleDir += bundlepDirPart

	addToBundle(bundleTempDir, filepath.Dir(bundle.Name()))

	filesMap := map[string]*ilmtv1.FileInfoCTLAction{}
	fileNames := map[string]interface{}{}

	for _, f := range files {
		filesMap[f.Name+f.Source+f.SourceType] = f
		fileNames[f.Name] = nil
	}

	currentMeteringExport.Files = make([]*ilmtv1.FileInfoCTLAction, 0, len(filesMap))

	for i := range filesMap {
		currentMeteringExport.Files = append(currentMeteringExport.Files, filesMap[i])
	}

	return len(files), nil
}

func makeArchive(source, target string) error {
	filename := filepath.Base(source)
	target = filepath.Join(target, fmt.Sprintf(TARGETGZ, filename))
	err := copyFile(source, target)
	if err != nil {
		return err
	}
	return nil
}

func addToBundle(source, target string) error {
	filename := filepath.Base(source)
	target = filepath.Join(target, fmt.Sprintf(TARGET, filename))
	err := copyFile(source, target)
	if err != nil {
		return err
	}
	return nil
}

func copyFile(source string, target string) error {
	tarfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	tarSphere := tar.NewWriter(tarfile)
	defer tarSphere.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	return filepath.Walk(source,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			if baseDir != EMPTY {
				header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
			}

			if err := tarSphere.WriteHeader(header); err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tarSphere, file)
			return err
		})
}
