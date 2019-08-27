// Copyright (c) 2019 devblok
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package kar

import (
	"io"
	"os"
)

// NewReader creates a reader from r. It will also check
// if the file is actually a kar archive, will return error
// when file incorrect.
func NewReader(r io.ReaderAt) (*Reader, error) {
	return &Reader{
		reader: r,
	}, nil
}

// Reader provides concurrent io for a kar file, and provides
// an os.File compatible layer to read files.
type Reader struct {
	// TODO: Implement a File like interface, so this archive and files could
	// potentially be used interchangeably. This means giving out handles from this package.
	reader io.ReaderAt
}

// ReadAll queries for the contents of given name
func (r *Reader) ReadAll(name string) ([]byte, error) {
	return []byte{}, nil
}

// Open opens a file on the archive and returns the imaginary handle.
func (r *Reader) Open(name string) (*File, error) {
	return &File{
		reader: r,
	}, nil
}

// File implements the same interface as os.File, for use in situations where a
// drop-in replacement for it would be useful. It behaves exactly like a file for reads and
// writes, but other operations for change of ownership and etc have no effect.
type File struct {
	os.File

	reader *Reader
}
