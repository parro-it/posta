package app

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/parro-it/posta/config"
	"github.com/parro-it/posta/plex"
)

type AppStarted struct{}

type Processor func(ctx context.Context) chan error

type App struct {
	Actions plex.Demux[any]
}

var Instance App

func (a *App) Start(ctx context.Context, processors ...Processor) {
	cfgpath, fail := config.GetCfgPath()
	if fail {
		return
	}
	config.Values.ConfigFile = cfgpath

	err := config.Init()
	if err != nil {
		log.Fatal(err)
	}

	a.Actions.Start()

	var errs plex.Mux[error]
	var cancels = make([]context.CancelFunc, len(processors))

	for idx, start := range processors {
		var procCtx context.Context
		procCtx, cancels[idx] = context.WithCancel(ctx)
		errs.AddInputFrom(start(procCtx))
	}
	go func() {
		for err := range errs.Output {
			proc := processors[err.Idx]
			cancelCtx := cancels[err.Idx]
			fmt.Fprintf(os.Stderr, "An error occurred in processor %v: %s. Processor will be restarted.", proc, err.Value)
			cancelCtx()

			var procCtx context.Context
			procCtx, cancels[err.Idx] = context.WithCancel(ctx)
			errs.AddInputFrom(proc(procCtx))
		}
	}()
	a.Actions.Input <- AppStarted{}
}

func PostAction(a any) {
	Instance.Actions.Input <- a
}

func ListenAction[T any]() chan T {
	return plex.AddOut[T](Instance.Actions)

}
