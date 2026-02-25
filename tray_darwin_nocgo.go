//go:build darwin && !cgo

package main

import "context"

func (a *App) SetupTray(ctx context.Context) {
}

func (a *App) updateTrayState(connected bool) {
}

func (a *App) OnExit() {
}
