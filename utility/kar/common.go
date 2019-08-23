// Copyright (c) 2019 devblok
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package kar

// FileHeader is the header at the start of the file.
// It will always be KAR\x00 followed by the size of
// the header that comes after it.
type FileHeader struct {
	Magic      [4]byte
	HeaderSize int32
}

// IndexEntry is info for one file in the file index.
type IndexEntry struct {
	Name           string
	Offset         int64
	Size           int64
	CompressedSize int64
}

// Header is the file header for kar files.
type Header struct {
	Author      string
	DateCreated int64
	Version     int64
	Index       []IndexEntry
}
