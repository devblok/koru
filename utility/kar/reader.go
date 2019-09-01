// Copyright (c) 2019 devblok
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package kar

import (
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/pierrec/lz4"
)

// Open opens the kar archived from r. It will also check
// if the file is actually a kar archive, will return an error
// when file incorrect.
func Open(r io.ReaderAt) (*Archive, error) {
	magic := make([]byte, MagicLength)
	if num, err := r.ReadAt(magic, 0); err != nil {
		return nil, err
	} else if num < MagicLength || strings.Compare(string(magic), "KAR\x00") != 0 {
		return nil, ErrFileFormat
	}

	headerSizeBytes := make([]byte, HeaderSizeNumberLength)
	if num, err := r.ReadAt(headerSizeBytes, MagicLength); err != nil {
		return nil, err
	} else if num < HeaderSizeNumberLength {
		return nil, ErrFileFormat
	}

	headerSize, err := binaryToint64(headerSizeBytes)
	if err != nil {
		return nil, ErrFileFormat
	}

	headerBytes := make([]byte, headerSize)
	if num, err := r.ReadAt(headerBytes, MagicLength+HeaderSizeNumberLength); err != nil {
		return nil, err
	} else if int64(num) < headerSize {
		return nil, ErrFileFormat
	}

	var header Header
	if err := gobDecode(&header, headerBytes); err != nil {
		return nil, err
	}
	return &Archive{
		reader: r,
		header: header,
	}, nil
}

// Archive provides concurrent io for a kar file, and can provide
// an io.Reader for each file separately to perform actions on.
type Archive struct {
	reader io.ReaderAt
	header Header
}

// GetFileInfo queries for a file with a given name in the archive
// and returns it's info if found. If not found it will return os.ErrNotExist error.
func (a *Archive) GetFileInfo(name string) (IndexEntry, error) {
	for _, f := range a.header.Index {
		if strings.Compare(name, f.Name) == 0 {
			return f, nil
		}
	}
	return IndexEntry{}, os.ErrNotExist
}

// ReadAll returns the entire contents of a file with a given name
func (a *Archive) ReadAll(name string) ([]byte, error) {
	e, err := a.GetFileInfo(name)
	if err != nil {
		return []byte{}, err
	}

	rawContents := make([]byte, e.CompressedSize)
	if n, err := a.reader.ReadAt(rawContents, e.Offset); err != nil {
		return []byte{}, err
	} else if int64(n) < e.CompressedSize {
		return []byte{}, ErrIOMisc
	}

	fileContents := make([]byte, e.Size)
	reader := lz4.NewReader(bytes.NewReader(rawContents))
	if n, err := reader.Read(fileContents); err != nil {
		return []byte{}, err
	} else if int64(n) < e.Size {
		return []byte{}, ErrIOMisc
	}

	return fileContents, nil
}

// Open returns a Reader for a file in the Archive
func (a *Archive) Open(name string) (*Reader, error) {
	e, err := a.GetFileInfo(name)
	if err != nil {
		return nil, err
	}

	sectionReader := io.NewSectionReader(a.reader, e.Offset, e.CompressedSize)
	return &Reader{
		reader: lz4.NewReader(sectionReader),
	}, nil
}

// Reader is a reader for a single file in an Archive.
// Abstracts away the location that needs to be known.
type Reader struct {
	io.Reader

	reader io.Reader
}

// Read reads already decompressed data
func (r *Reader) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}
