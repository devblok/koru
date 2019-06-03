package main

import (
	"errors"

	"github.com/gobuffalo/packr"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	log "github.com/sirupsen/logrus"
)

// Global variables for GTK and resources
var (
	Builder         *gtk.Builder
	StaticResources packr.Box
)

func init() {
	StaticResources = packr.NewBox("./resources")
}

func buildInterface() (*gtk.Application, error) {
	app, err := gtk.ApplicationNew("org.koru3d.korued", glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		return nil, err
	}

	app.Connect("startup", func() {
		log.Info("Application starting")
	})

	app.Connect("activate", func() {
		log.Info("Application activating")

		resource, err := StaticResources.FindString("korued.glade")
		if err != nil {
			log.Fatal(err)
			panic(err)
		}

		builder, err := gtk.BuilderNew()
		builder.AddFromString(resource)
		if err != nil {
			log.Error(err)
			panic(err)
		}

		Builder = builder

		obj, err := builder.GetObject("mainWindow")
		if err != nil {
			log.Error(err)
		}

		var (
			ok  bool
			win *gtk.Window
		)

		if win, ok = obj.(*gtk.Window); !ok {
			log.Error(errors.New("failed to cast Object from builder to Window"))
		} else {
			win.SetDefaultSize(600, 480)

			win.ShowAll()
			app.AddWindow(win)
		}
	})

	app.Connect("shutdown", func() {
		log.Info("Application shutting down")
	})
	return app, nil
}
