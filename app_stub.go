//go:build !linux && !windows

package main

import (
	"context"
	"os/exec"
)

func (a *App) configureCmd(cmd *exec.Cmd) {
}

func (a *App) EnableAutostart() error {
	return nil
}

func (a *App) DisableAutostart() error {
	return nil
}

func (a *App) platformInit() error {
	return nil
}

func (a *App) cleanupZombies() {
}

func (a *App) SetupTray(ctx context.Context) {
}

func (a *App) updateTrayState(connected bool) {
}