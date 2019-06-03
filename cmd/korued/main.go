package main

import (
	"os"

	"github.com/gotk3/gotk3/gtk"
	log "github.com/sirupsen/logrus"
)

var Builder *gtk.Builder

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
