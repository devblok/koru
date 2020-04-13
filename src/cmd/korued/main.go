// Copyright (c) 2019 devblok
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"os"

	"github.com/gotk3/gotk3/gtk"
	log "github.com/sirupsen/logrus"
)

func init() {
	gtk.Init(&os.Args)
}

func main() {
	app, err := buildInterface()
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	os.Exit(app.Run(os.Args))
}
