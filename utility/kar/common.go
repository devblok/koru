// Copyright (c) 2019 devblok
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package kar

import (
	"bytes"
	"encoding/binary"
)

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

// MaxExpectedSize calculates the amount of space a Header could take.
// It's important to know this before writing the header into the file.
// It only needs to be roughtly correct, offsets will be calculated
// with consideration for this number
func (h *Header) MaxExpectedSize() int64 {
	var size int64
	size += int64(len(h.Author))
	size += 16 // DataCreated + Version
	size += 60 // Names etc
	for _, e := range h.Index {
		size += int64(len(e.Name))
		size += 24 // numbers
		size += 60
	}
	return size
}

func int64ToBinary(num int64) []byte {
	buf := bytes.NewBuffer([]byte{})
	if err := binary.Write(buf, binary.LittleEndian, &num); err != nil {
		panic(err) // If this thing fails you're probably having bigger problems
	}
	return buf.Bytes()
}

func binaryToint64(bts []byte) (int64, error) {
	var num int64
	if err := binary.Read(bytes.NewReader(bts), binary.LittleEndian, &num); err != nil {
		return 0, err
	}
	return num, nil
}
