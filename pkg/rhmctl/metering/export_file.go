package metering

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"emperror.dev/errors"
	"github.com/redhat-marketplace/rhmctl/pkg/rhmctl/config"
)

type BundleFile struct {
	file *os.File
	tar  *tar.Writer
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

func (f *BundleFile) open(filepath string) error {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, fileMode)
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
