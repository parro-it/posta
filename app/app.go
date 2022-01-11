package app

import (
	"log"

	"github.com/parro-it/posta/config"
	"github.com/parro-it/posta/plex"
)

type AppStarted struct{}

type App struct {
	Actions plex.Demux[any]
}

var Instance App

func (a App) Start() {
	cfgpath, fail := config.GetCfgPath()
	if fail {
		return
	}
	config.Values.ConfigFile = cfgpath

	err := config.Init()
	if err != nil {
		log.Fatal(err)
	}

	a.Actions.Input <- AppStarted{}

}
