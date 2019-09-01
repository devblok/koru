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
	"log"
	"math"
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
		log.Println(err)
		return nil, ErrTempFail
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
func (b *Builder) Add(name string, r io.Reader) error {
	tempName := strconv.Itoa(time.Now().Nanosecond())
	f, err := os.Create(filepath.Join(b.tempDir, tempName))
	if err != nil {
		log.Println(err)
		return ErrTempFail
	}
	defer f.Close()
	writer := lz4.NewWriter(f)
	written, err := io.Copy(writer, r)
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
// into a kar archive that is ready to use. This function may block for
// a long time, cosidering it does a lot of operations. It will keep the mutex locked.
func (b *Builder) WriteTo(w io.Writer) (int64, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Fill up the index so we could calculate the approximate size it will take
	header := b.header
	for _, v := range b.files {
		header.Index = append(header.Index, IndexEntry{
			Name:           v.Name,
			Size:           v.Size,
			CompressedSize: v.Compressed,
			Offset:         math.MinInt64,
		})
	}

	// Write the magic letters and the approximate size of the header
	var offset int64
	magicSize, err := w.Write([]byte("KAR\x00"))
	if err != nil {
		return 0, err
	}
	headerBytesSize := header.MaxExpectedSize()
	headerSizeBytesWritten, err := w.Write(int64ToBinary(headerBytesSize))
	if err != nil {
		return 0, err
	}

	// Always ensure HeaderSizeNumberLength amount of bytes is written
	if headerSizeBytesWritten < HeaderSizeNumberLength {
		if _, err := w.Write(make([]byte, HeaderSizeNumberLength-headerSizeBytesWritten)); err != nil {
			return 0, err
		}
	}

	// the offset at which we start writing files
	offset += int64(magicSize)
	offset += int64(HeaderSizeNumberLength)
	offset += headerBytesSize

	// figure out files offsets
	for idx := range header.Index {
		header.Index[idx].Offset = offset
		offset += header.Index[idx].CompressedSize
	}

	// encode and write completed header
	rawHeader, err := gobEncode(header)
	if err != nil {
		return 0, err
	}
	w.Write(rawHeader)

	// write the padding
	w.Write(make([]byte, int(headerBytesSize)-len(rawHeader)))

	// write out all the files
	// order should be preserved beforehand,
	// otherwise we will write incorrectly.
	// We don't increment offset, because it's precalculated
	for _, file := range b.files {
		f, err := os.Open(filepath.Join(b.tempDir, file.TempName))
		if err != nil {
			log.Println(err)
			return 0, ErrTempFail
		}
		if _, err := io.Copy(w, f); err != nil {
			return 0, err
		}
		f.Close()
	}

	// delete the index
	b.files = b.files[:0]
	return offset, nil
}

func gobEncode(data interface{}) ([]byte, error) {
	var encoded bytes.Buffer
	enc := gob.NewEncoder(&encoded)
	if err := enc.Encode(data); err != nil {
		return nil, err
	}
	return encoded.Bytes(), nil
}

func gobDecode(obj interface{}, bts []byte) error {
	dec := gob.NewDecoder(bytes.NewBuffer(bts))
	if err := dec.Decode(obj); err != nil {
		return err
	}
	return nil
}
