// Copyright (c) 2019 devblok
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"errors"
	"flag"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/devblok/koru/src/utility/kar"
)

func init() {
	u, err := user.Current()
	if err != nil {
		currentUserName = "unknown"
	}
	currentUserName = u.Name
}

var (
	currentUserName string
	author          = flag.String("author", currentUserName, "Set the author of the package when compressing")
	version         = flag.Int64("version", 1, "Archive version number to create it with")
	extract         = flag.String("e", "", "Extract the file given")
	compress        = flag.String("c", "", "Compress the given file/folder")
	dstFile         = flag.String("f", "out.kar", "Destination file")
	silent          = flag.Bool("s", false, "Silent")
)

func main() {
	var opMade bool
	flag.Parse()

	if *extract != "" && *compress != "" {
		panic(errors.New("only one operation at a time"))
	}

	if *extract != "" {
		opMade = true
		panic("not implemented")
	}

	if *compress != "" {
		opMade = true
		if err := compressFiles(); err != nil {
			panic(err)
		}
	}

	if !opMade {
		flag.PrintDefaults()
	}
}

func compressFiles() error {
	if _, err := os.Stat(*dstFile); err == nil {
		return errors.New("destination file exists, will not overwrite")
	}

	dst, err := os.Create(*dstFile)
	if err != nil {
		return err
	}

	var filesToCompress []string
	filepath.Walk(*compress, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		filesToCompress = append(filesToCompress, path)
		return nil
	})

	karBuilder, err := kar.NewBuilder(kar.Header{
		Author:      currentUserName,
		DateCreated: time.Now().Unix(),
		Version:     *version,
	})
	if err != nil {
		return err
	}

	for _, ftc := range filesToCompress {
		f, err := os.Open(ftc)
		if err != nil {
			return err
		}
		karBuilder.Add(ftc, f)
	}

	karBuilder.WriteTo(dst)
	return nil
}
