package app

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gotk3/gotk3/gdk"
	"github.com/parro-it/posta/chans"
	"github.com/parro-it/posta/config"
)

type AppStarted struct{}

type Processor func(ctx context.Context) chan error

type App struct {
	actions   chans.Demux[any]
	shortcuts map[uint64]Shortcut
}

type Shortcut func() error

func (a *App) RegisterShortcut(key uint32, state uint32, exec Shortcut) {
	a.shortcuts[uint64(key)|uint64(state)<<32] = exec
}

var Instance App

type KeyPressed struct {
	Key   uint
	State uint
}

func (a *App) PostKeyPressed(key uint, state uint) {
	PostAction(KeyPressed{
		Key:   key,
		State: state,
	})
	for _, mask := range []uint{gdk.CONTROL_MASK, gdk.META_MASK, uint(gdk.SHIFT_MASK)} {
		if state&mask == mask {
			if sh, ok := a.shortcuts[uint64(key)|uint64(mask)<<32]; ok {
				err := sh()
				if err != nil {
					panic(err)
				}
				break
			}
		}
	}

}

func (a *App) Start(ctx context.Context, processors ...Processor) {
	a.shortcuts = map[uint64]Shortcut{}
	cfgpath, fail := config.GetCfgPath()
	if fail {
		return
	}
	config.Values.ConfigFile = cfgpath

	err := config.Init()
	if err != nil {
		log.Fatal(err)
	}

	a.actions.Start()

	var errs chans.Mux[error]
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
			fmt.Fprintf(os.Stderr, "An error occurred in processor %v: %s. Processor will be restarted.", proc, err.Value.Error())
			cancelCtx()

			var procCtx context.Context
			procCtx, cancels[err.Idx] = context.WithCancel(ctx)
			errs.AddInputFrom(proc(procCtx))
		}
	}()
	a.actions.Input <- AppStarted{}
}

func PostAction(a any) {
	Instance.actions.Input <- a
}

func ListenAction[T any]() chan T {
	return chans.AddOut[T](Instance.actions)
}
func ListenAction2[T1 any, T2 any]() chan any {
	return chans.AddOut2[T1, T2](Instance.actions)
}
