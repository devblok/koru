// Copyright (c) 2019 devblok
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

// Package kar is an api for an lz4 backed file format.
// It's purpose is to be well suited for resource streaming resources
// from it. It's designed to be memory mapped, so (unlike tar) it knows
// where all the files are located before they're read. This nescesitates
// a bit of an unusual setup, where the archive itself is not compressed in
// any form, rather every file is individually compressed, so it could be immediately
// read from it's place and decompressed on the fly. This somewhat compromises
// space efficiency, but space efficiency is not the primary goal of this
// package. It instead focuses on getting resources from disk to a usable
// state as fast as possible. It can be read from concurrently.
package kar

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
)

// package errors
var (
	ErrFileFormat = errors.New("corrupted or not a kar archive")
	ErrTempFail   = errors.New("temporary folder or file operation failed")
	ErrIOMisc     = errors.New("some unknown error unhandled by the io occured")
)

// Sizes relevant to the header of file
const (
	MagicLength            = 4
	HeaderSizeNumberLength = 16
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
