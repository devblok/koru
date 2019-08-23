// Copyright (c) 2019 devblok
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package kar

import "io"

// NewReaderAt creates a reader from r.
func NewReaderAt(r io.ReaderAt) *ReaderAt {
	return nil
}

// ReaderAt provides concurrent io for a kar file.
type ReaderAt struct {
	io.ReaderAt
}

// ReadAt reads the file at that location
func (ReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, nil
}
