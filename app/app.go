package app

import (
	"log"

	"github.com/parro-it/posta/actions"
	"github.com/parro-it/posta/config"
)

type AppStarted struct{}

const APP_STARTED = 2

func (a AppStarted) Type() actions.ActionType {
	return APP_STARTED
}

type App struct {
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

	actions.Post(AppStarted{})

}
