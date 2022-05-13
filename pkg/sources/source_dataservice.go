package sources

import (
	"context"
	"time"

	"github.com/gotidy/ptr"
	"github.com/redhat-marketplace/datactl/pkg/bundle"
	"github.com/redhat-marketplace/datactl/pkg/clients/dataservice"
	"github.com/redhat-marketplace/datactl/pkg/datactl/api"
	dataservicev1 "github.com/redhat-marketplace/datactl/pkg/datactl/api/dataservice/v1"
	"github.com/redhat-marketplace/datactl/pkg/printers"
)

type dataServiceSource struct {
	printers.TablePrinter
	dataService dataservice.Client
}

type dataServiceOptions struct {
	GenericOptions
}

const (
	IncludeDeleted string = "includeDeleted"
	BeforeDate            = "beforeDate"
	AfterDate             = "afterDate"
	DryRun                = "dryRun"
)

func NewDataServiceOptions(includeDeleted bool, beforeDate, afterDate time.Time, dryRun bool) GenericOptions {
	return NewOptions(
		IncludeDeleted, includeDeleted,
		BeforeDate, beforeDate,
		AfterDate, afterDate,
		DryRun, dryRun,
	)
}

func NewDataService(
	dataService dataservice.Client,
	printer printers.TablePrinter,
) (CommitableSource, error) {
	d := &dataServiceSource{
		dataService:  dataService,
		TablePrinter: printer,
	}

	return d, nil
}

func (d *dataServiceSource) Commit(
	ctx context.Context,
	currentMeteringExport *api.MeteringExport,
	bundle *bundle.BundleFile,
	opts GenericOptions) error {

	dryRun, _, err := opts.GetBool(DryRun)
	if err != nil {
		return err
	}

	errs := map[string]error{}
	committed := 0

	for _, file := range currentMeteringExport.Files {
		file.Action = dataservicev1.Commit
		file.Result = dataservicev1.Ok

		if dryRun || file.Committed == true {
			if dryRun {
				file.Result = dataservicev1.DryRun
			}

			d.TableOutput(func(po printers.PrintObj) {
				po.Print(file)
			})

			continue
		}

		err := d.dataService.DeleteFile(ctx, file.Id)
		if err != nil {
			file.Error = err.Error()
			file.Committed = false
			file.Result = dataservicev1.Error

			d.TableOutput(func(po printers.PrintObj) {
				po.Print(file)
			})

			errs[file.Name] = err
			continue
		}

		file.Error = ""
		file.Committed = true
		committed = committed + 1

		d.TableOutput(func(po printers.PrintObj) {
			po.Print(file)
		})
	}

	err = bundle.Close()
	if err != nil {
		return err
	}

	err = bundle.Compact(nil)
	if err != nil {
		return err
	}

	if dryRun {
		return nil
	}

	return nil
}

func (d *dataServiceSource) Pull(
	ctx context.Context,
	currentMeteringExport *api.MeteringExport,
	bundle *bundle.BundleFile,
	options GenericOptions,
) error {
	includeDeleted, _, err := options.GetBool(IncludeDeleted)
	if err != nil {
		return err
	}

	beforeDate, _, err := options.GetTime(BeforeDate)
	if err != nil {
		return err
	}

	afterDate, _, err := options.GetTime(AfterDate)
	if err != nil {
		return err
	}

	response := dataservicev1.ListFilesResponse{}
	listOpts := dataservice.ListOptions{
		IncludeDeleted: includeDeleted,
		BeforeDate:     beforeDate,
		AfterDate:      afterDate,
	}

	files := []*dataservicev1.FileInfoCTLAction{}
	errs := map[string]error{}
	found := 0
	pulled := 0

	for {
		err := d.dataService.ListFiles(ctx, listOpts, &response)

		if err != nil {
			return err
		}

		for i := range response.Files {
			cliFile := dataservicev1.NewFileInfoCTLAction(response.Files[i])
			files = append(files, cliFile)
			found = found + 1

			w, err := bundle.NewFile(cliFile.Name, int64(cliFile.Size))
			if err != nil {
				return err
			}

			_, err = d.dataService.DownloadFile(ctx, cliFile.Id, w)
			if err != nil {
				cliFile.Action = dataservicev1.Pull
				cliFile.Result = dataservicev1.Error
				cliFile.Error = err.Error()
				errs[cliFile.Name] = err

				d.TableOutput(func(tosp printers.PrintObj) {
					tosp.Print(cliFile)
				})
				continue
			}

			cliFile.Action = dataservicev1.Pull
			cliFile.Result = dataservicev1.Ok
			pulled = pulled + 1

			d.TableOutput(func(tosp printers.PrintObj) {
				tosp.Print(cliFile)
			})
		}

		if response.NextPageToken == "" {
			break
		}

		listOpts.PageSize = ptr.Int(int(response.PageSize))
		listOpts.PageToken = response.NextPageToken
	}

	filesMap := map[string]*dataservicev1.FileInfoCTLAction{}
	fileNames := map[string]interface{}{}

	for _, f := range currentMeteringExport.Files {
		if f.Committed && f.Pushed {
			continue
		}

		filesMap[f.Name+f.Source+f.SourceType] = f
		fileNames[f.Name] = nil
	}

	for _, f := range files {
		filesMap[f.Name+f.Source+f.SourceType] = f
		fileNames[f.Name] = nil
	}

	currentMeteringExport.Files = make([]*dataservicev1.FileInfoCTLAction, 0, len(filesMap))

	for i := range filesMap {
		currentMeteringExport.Files = append(currentMeteringExport.Files, filesMap[i])
	}

	return nil
}
