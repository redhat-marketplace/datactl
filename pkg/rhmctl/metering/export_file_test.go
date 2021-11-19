package metering

import (
	"archive/tar"
	"bytes"
	"io"
	"io/fs"
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("export_file", func() {
	It("should create a tar", func() {
		tmpdir := os.TempDir()
		file, err := ioutil.TempFile(tmpdir, "export_file_*.tar")
		Expect(err).To(Succeed())

		tarFile, err := NewBundle(file.Name())
		Expect(err).To(Succeed())

		lens := []int64{}

		buf := []byte("this is a string")
		lens = append(lens, int64(len(buf)))
		w, err := tarFile.NewFile("test", int64(len(buf)))
		Expect(err).To(Succeed())
		io.Copy(w, bytes.NewReader(buf))

		buf = []byte("this is another string")
		lens = append(lens, int64(len(buf)))
		w, err = tarFile.NewFile("test2", int64(len(buf)))
		Expect(err).To(Succeed())
		io.Copy(w, bytes.NewReader(buf))

		Expect(tarFile.Close()).To(Succeed())

		tarFile, err = NewBundle(file.Name())
		Expect(err).To(Succeed())

		buf = []byte("this is a string 3")
		lens = append(lens, int64(len(buf)))
		w, err = tarFile.NewFile("test3", int64(len(buf)))
		Expect(err).To(Succeed())
		io.Copy(w, bytes.NewReader(buf))

		buf = []byte("this is another string 4")
		lens = append(lens, int64(len(buf)))
		w, err = tarFile.NewFile("test4", int64(len(buf)))
		Expect(err).To(Succeed())
		io.Copy(w, bytes.NewReader(buf))

		Expect(tarFile.Close()).To(Succeed())

		f, err := os.OpenFile(file.Name(), os.O_RDONLY, fileMode)
		tarReader := tar.NewReader(f)
		files := []fs.FileInfo{}

		for {
			header, err := tarReader.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				Expect(err).To(Succeed())
			}
			files = append(files, header.FileInfo())
		}

		expectedFiles := []string{"test", "test2", "test3", "test4"}

		Expect(files).To(HaveLen(4))
		for i := 0; i < len(files); i++ {
			file := files[i]
			Expect(file.Name()).To(Equal(expectedFiles[i]))
			Expect(file.Size()).To(Equal(lens[i]))
		}

		os.RemoveAll(tmpdir)
		tarFile, err = NewBundle(file.Name())
		Expect(err).To(Succeed())
	})
})
