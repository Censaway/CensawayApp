//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#include <stdlib.h>
#include "tray_impl_darwin.h"
*/
import "C"

import (
	"context"
	"unsafe"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var trayCtx context.Context
var trayApp *App

const (
	ID_STATUS  = 0
	ID_CONNECT = 1
	ID_SHOW    = 2
	ID_QUIT    = 3
)

func (a *App) SetupTray(ctx context.Context) {
	trayCtx = ctx
	trayApp = a

	if len(a.Icon) > 0 {
		cData := unsafe.Pointer(&a.Icon[0])
		cLen := C.int(len(a.Icon))
		C.init_tray_c(cData, cLen)
	}
}

func (a *App) OnExit() {
}

func (a *App) updateTrayState(connected bool) {
	if connected {
		C.update_tray_state_c(1)
	} else {
		C.update_tray_state_c(0)
	}
}

//export goOnTrayClick
func goOnTrayClick(id C.int) {
	switch int(id) {
	case ID_CONNECT:
		if trayApp.GetRunningState() {
			trayApp.StopVless()
		} else {
			if trayApp.Settings.LastProfileID != "" {
				trayApp.connectLastProfile()
			} else {
				wailsRuntime.WindowShow(trayCtx)
			}
		}
	case ID_SHOW:
		wailsRuntime.WindowShow(trayCtx)
		wailsRuntime.WindowUnminimise(trayCtx)
	case ID_QUIT:
		trayApp.isQuitting = true
		trayApp.StopVless()
		trayApp.shutdownWg.Wait()
		trayApp.cleanupZombies()
		wailsRuntime.Quit(trayCtx)
	}
}