// Copyright (c) 2019 devblok
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package kar

import (
	"bytes"
	"encoding/gob"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/pierrec/lz4"
)

// NewBuilder creates a new Builder. Do not fill the Index in
// the header, it will be overwritten anyway.
func NewBuilder(header Header) (*Builder, error) {
	temp, err := ioutil.TempDir("", "karBuilder")
	if err != nil {
		return nil, err
	}
	builder := &Builder{
		tempDir: temp,
		header:  header,
	}
	// TODO: Not sure if this is a good place to clean up.
	// Measure if GC will take a hit later.
	runtime.SetFinalizer(builder, func(builder *Builder) {
		os.RemoveAll(builder.tempDir)
	})
	return builder, nil
}

type tempFile struct {

	// Name is the actual name of the file
	Name string

	// TempName is the temporary name given by the Builder
	TempName string

	// Size in compressed state
	Size int64

	Compressed int64
}

// Builder is the high level builder for the archive format.
// Arhives are versioned and cannot be appended to, This Builder
// is the way to create an archive. Whenever Add is called, KarBuilder
// will create a temporary dir, where it will store compressed files,
// then finally bundling them togeter and writing them out with WriteTo.
type Builder struct {
	io.WriterTo

	tempDir string
	header  Header

	mutex sync.Mutex
	files []tempFile
}

// Add appends data to the builder with a given name.
// Will block until lz4 finishes compression. Is safe
// to use concurrently in different goroutines.
func (b *Builder) Add(name string, data []byte) error {
	tempName := strconv.Itoa(time.Now().Nanosecond())
	f, err := os.Create(filepath.Join(b.tempDir, tempName))
	if err != nil {
		return err
	}
	defer f.Close()
	reader := bytes.NewReader(data)
	writer := lz4.NewWriter(f)
	written, err := io.Copy(writer, reader)
	if err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	info, err := f.Stat()
	if err != nil {
		return err
	}
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.files = append(b.files, tempFile{
		Name:       name,
		TempName:   tempName,
		Size:       written,
		Compressed: info.Size(),
	})
	return nil
}

// WriteTo bundles and writes all of the files added to the Builder
// into a kar archive that is ready to use.
func (b *Builder) WriteTo(w io.Writer) (int64, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	header := b.header
	for _, v := range b.files {
		header.Index = append(header.Index, IndexEntry{
			Name:           v.Name,
			Size:           v.Size,
			CompressedSize: v.Compressed,
			Offset:         0,
		})
	}

	rawHeader, err := encode(header)
	if err != nil {
		return 0, err
	}

	fileHeader := FileHeader{
		Magic:      [...]byte{'K', 'A', 'R', '\x00'},
		HeaderSize: int32(len(rawHeader)),
	}

	rawFileHeader, err := encode(fileHeader)
	if err != nil {
		return 0, err
	}

	w.Write(rawFileHeader)
	w.Write(rawHeader)

	b.files = b.files[:0]
	return 0, nil
}

func encode(data interface{}) ([]byte, error) {
	var encoded bytes.Buffer
	enc := gob.NewEncoder(&encoded)
	if err := enc.Encode(data); err != nil {
		return nil, err
	}
	return encoded.Bytes(), nil
}
