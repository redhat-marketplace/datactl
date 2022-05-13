package sources

import (
	"context"
	"fmt"
	"time"

	"github.com/redhat-marketplace/datactl/pkg/bundle"
	"github.com/redhat-marketplace/datactl/pkg/datactl/api"
	"github.com/redhat-marketplace/datactl/pkg/datactl/config"
	"github.com/redhat-marketplace/datactl/pkg/printers"
)

// Goals:
// * Each source will be able to be configured independently.
// * We'll need a config for each source.

// A pull typically
// a. Queries data - in the case of dataservice is all files available.
//    For the case of ILMT is endpoint data.
// b. Saves the data in the required format. Pull should go ahead and transform?
// c. Accepts a bundle to write the files to.
// d. For sources other than data service, we'll need to save a checkpoint to continue from.

type CommitableSource interface {
	Source
	Commit(
		ctx context.Context,
		currentMeteringExport *api.MeteringExport,
		bundle *bundle.BundleFile,
		opts GenericOptions,
	) error
}

type Source interface {
	Pull(
		ctx context.Context,
		currentMeteringExport *api.MeteringExport,
		bundle *bundle.BundleFile,
		options GenericOptions,
	) error
}

type Factory interface {
	FromSource(source api.Source) (Source, error)
}

type sourceFactory struct {
	rhmConfigFlags *config.ConfigFlags
	printer        printers.Printer
}

func (s *sourceFactory) FromSource(source api.Source) (Source, error) {
	switch source.Type {
	case api.DataService:
		dataService, err := s.rhmConfigFlags.DataServiceClient(source)
		if err != nil {
			return nil, err
		}

		return NewDataService(dataService, s.printer)
	}

	return nil, fmt.Errorf("sourceType %s not found", source.Type)
}

type SourceFactoryBuilder struct {
	rhmConfigFlags *config.ConfigFlags
	printer        printers.Printer

	errs []error
}

func (s *SourceFactoryBuilder) SetPrinter(
	p printers.Printer,
) *SourceFactoryBuilder {
	s.printer = p
	return s
}

func (s *SourceFactoryBuilder) SetConfigFlags(rhmConfigFlags *config.ConfigFlags) *SourceFactoryBuilder {
	s.rhmConfigFlags = rhmConfigFlags
	return s
}

func (s *SourceFactoryBuilder) Build() Factory {
	return &sourceFactory{
		rhmConfigFlags: s.rhmConfigFlags,
		printer:        s.printer,
	}
}

func EmptyOptions() GenericOptions {
	return &Options{opts: make(map[string]interface{})}
}

type GenericOptions interface {
	GetString(string) (string, bool, error)
	GetInt(string) (int, bool, error)
	GetBool(string) (bool, bool, error)
	GetTime(string) (time.Time, bool, error)
	Get(string) (interface{}, bool, error)
}

type Options struct {
	opts map[string]interface{}
}

func NewOptions(key string, value interface{}, fields ...interface{}) GenericOptions {
	opts := make(map[string]interface{})

	opts[key] = value

	if len(fields) > 0 && len(fields)%2 == 0 {
		for i := 0; i < len(fields); i = i + 2 {
			key, value := fields[i], fields[i+1]

			if k, ok := key.(string); ok {
				opts[k] = value
			}
			if k, ok := key.(fmt.Stringer); ok {
				opts[k.String()] = value
			}
		}
	}

	return &Options{opts: opts}
}

func (o *Options) GetString(name string) (string, bool, error) {
	if v, ok := o.opts[name]; ok {
		s, ok := v.(string)
		if !ok {
			return "", false, fmt.Errorf("failed to convert type %t to string", v)
		}

		return s, true, nil
	}

	return "", false, nil
}

func (o *Options) GetInt(name string) (int, bool, error) {
	if v, ok := o.opts[name]; ok {
		s, ok := v.(int)
		if !ok {
			return 0, false, fmt.Errorf("failed to convert type %t to int", v)
		}

		return s, true, nil
	}

	return 0, false, nil
}

func (o *Options) GetBool(name string) (bool, bool, error) {
	if v, ok := o.opts[name]; ok {
		s, ok := v.(bool)
		if !ok {
			return false, false, fmt.Errorf("failed to convert type %t to bool", v)
		}

		return s, true, nil
	}

	return false, false, fmt.Errorf("not found")
}

func (o *Options) GetTime(name string) (time.Time, bool, error) {
	if v, ok := o.opts[name]; ok {
		s, ok := v.(time.Time)
		if !ok {
			return time.Time{}, false, fmt.Errorf("failed to convert type %t to time.Time", v)
		}

		return s, true, nil
	}

	return time.Time{}, false, nil
}

func (o *Options) Get(name string) (interface{}, bool, error) {
	if v, ok := o.opts[name]; ok {
		return v, true, nil
	}

	return "", false, nil
}
