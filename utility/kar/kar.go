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

import "errors"

// package errors
var (
	ErrFileFormat = errors.New("corrupted or not a kar archive")
	ErrTempFail   = errors.New("temporary folder or file operation failed")
)

const (
	MagicLength            = 4
	HeaderSizeNumberLength = 8
)
