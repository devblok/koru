// Copyright (c) 2019 devblok
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package kar

import (
	"io"
	"strings"
)

// Open opens the kar archived from r. It will also check
// if the file is actually a kar archive, will return an error
// when file incorrect.
func Open(r io.ReaderAt) (*Archive, error) {
	ar := Archive{
		reader: r,
	}

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

	return &ar, nil
}

// Archive provides concurrent io for a kar file, and can provide
// an io.Reader for each file separately to perform actions on.
type Archive struct {
	reader io.ReaderAt
}

// ReadAll returns the entire contents of a file with a given name
func (a *Archive) ReadAll(file string) ([]byte, error) {
	return []byte{}, nil
}

// Open returns a Reader for a file in the Archive
func (a *Archive) Open(name string) (*Reader, error) {
	return &Reader{
		archive: a,
	}, nil
}

// Reader is a reader for a single file in an Archive.
// Abstracts away the location that needs to be known.
type Reader struct {
	io.Reader

	archive *Archive
}

// Read reads already decompressed data
func (r *Reader) Read(p []byte) (n int, err error) {
	return 0, nil
}
