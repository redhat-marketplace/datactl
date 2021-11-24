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

package metering

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"emperror.dev/errors"
	datactlapi "github.com/redhat-marketplace/datactl/pkg/datactl/api"
	"github.com/redhat-marketplace/datactl/pkg/datactl/config"
)

type BundleFile struct {
	file      *os.File
	tar       *tar.Writer
	tarReader *tar.Reader
}

var (
	_ io.Closer = &BundleFile{}
)

func NewBundle(filepath string) (b *BundleFile, err error) {
	b = &BundleFile{}
	err = b.open(filepath)
	return
}

const (
	fileMode os.FileMode = 0640
)

func (f *BundleFile) open(fileName string) error {
	dir := filepath.Dir(fileName)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, fileMode)
	if err != nil {
		return err
	}

	info, err := file.Stat()

	// if the tar file has been written to previously, we need to remove the last
	// 1024 bytes to append new files
	if info.Size() > 1024 {
		if _, err = file.Seek(-1024, os.SEEK_END); err != nil {
			return err
		}
	}

	f.file = file
	f.tar = tar.NewWriter(file)
	f.tarReader = tar.NewReader(file)

	return nil
}

func (f *BundleFile) Name() string {
	return f.file.Name()
}

func (f *BundleFile) NewFile(filename string, size int64) (io.Writer, error) {
	hdr := &tar.Header{
		Name: filename,
		Mode: int64(fileMode),
		Size: size,
	}

	if err := f.tar.WriteHeader(hdr); err != nil {
		return nil, err
	}
	return f.tar, nil
}

func (f *BundleFile) Close() error {
	return errors.Combine(f.tar.Close(), f.file.Close())
}

func (f *BundleFile) Walk(walk func(header *tar.Header, r io.Reader)) error {
	for {
		header, err := f.tarReader.Next()
		if err != nil && err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		walk(header, f.tarReader)
	}
	return nil
}

func (f *BundleFile) Compact(fileNames map[string]interface{}) error {
	headers := map[string]int{}
	os.Remove(f.Name() + "compact")
	newBundle, err := NewBundle(f.Name() + "compact")
	defer newBundle.Close()

	if err != nil {
		return err
	}

	var i int
	err = WalkTar(f.file.Name(), func(header *tar.Header, r io.Reader) error {
		headers[header.Name] = i
		i = i + 1
		return nil
	})

	if err != nil {
		return err
	}

	headersIntMap := map[int]interface{}{}
	for _, i := range headers {
		headersIntMap[i] = nil
	}

	i = 0
	err = WalkTar(f.file.Name(), func(header *tar.Header, r io.Reader) error {
		_, ok := headersIntMap[i]
		i = i + 1

		if !ok {
			return nil
		}

		if fileNames != nil {
			_, ok := fileNames[header.Name]

			if !ok {
				return nil
			}
		}

		var w io.Writer
		w, err = newBundle.NewFile(header.Name, header.Size)
		if err != nil {
			return err
		}

		_, err = io.Copy(w, r)
		if err != nil {
			return err
		}

		return nil
	})

	err = newBundle.Close()
	if err != nil {
		return err
	}

	return os.Rename(newBundle.Name(), f.Name())
}

func NewBundleWithDefaultName() (*BundleFile, error) {
	timestamp := time.Now().Format("20060102T150405Z")
	filename := filepath.Join(config.RecommendedDataDir, fmt.Sprintf("rhm-upload-%s.tar", timestamp))

	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	return NewBundle(filename)
}

func NewBundleFromExport(export *datactlapi.MeteringExport) (*BundleFile, error) {
	if export == nil {
		return nil, errors.New("export is nil")
	}

	if export.FileName == "" {
		bundle, err := NewBundleWithDefaultName()
		if err != nil {
			return nil, err
		}
		export.FileName = bundle.file.Name()
		return bundle, err
	}

	return NewBundle(export.FileName)
}

func WalkTar(filepath string, walk func(header *tar.Header, r io.Reader) error) error {
	file, err := os.OpenFile(filepath, os.O_RDONLY, fileMode)

	if err != nil {
		return err
	}

	defer file.Close()

	tarReader := tar.NewReader(file)

	for {
		header, err := tarReader.Next()
		if err != nil && err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		err = walk(header, tarReader)
		if err != nil && err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}
