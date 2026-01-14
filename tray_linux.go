//go:build linux

package main

import (
	"context"
	"time"

	"github.com/getlantern/systray"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	mConnect *systray.MenuItem
	mStatus  *systray.MenuItem
)

func (a *App) SetupTray(ctx context.Context) {
	go systray.Run(func() { a.onTrayReady(ctx) }, a.onTrayExit)
}

func (a *App) onTrayReady(ctx context.Context) {
	if len(a.Icon) > 0 {
		systray.SetIcon(a.Icon)
	} else {
		systray.SetTitle("VPN")
	}

	systray.SetTitle("Censaway")
	systray.SetTooltip("Censaway VPN Client")

	mStatus = systray.AddMenuItem("Status: Disconnected", "Current VPN status")
	mStatus.Disable()

	systray.AddSeparator()

	mConnect = systray.AddMenuItem("Connect", "Toggle VPN connection")
	
	mShow := systray.AddMenuItem("Show Window", "Show main window")

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Exit app")

	go func() {
		for {
			select {
			case <-mConnect.ClickedCh:
				if a.GetRunningState() {
					a.StopVless()
				} else {
					if a.Settings.LastProfileID != "" {
						a.connectLastProfile()
					} else {
						wailsRuntime.WindowShow(ctx)
						wailsRuntime.WindowUnminimise(ctx)
					}
				}
			case <-mShow.ClickedCh:
				wailsRuntime.WindowShow(ctx)
				wailsRuntime.WindowUnminimise(ctx)
				wailsRuntime.WindowSetAlwaysOnTop(ctx, true)
				time.Sleep(100 * time.Millisecond)
				wailsRuntime.WindowSetAlwaysOnTop(ctx, false)
			case <-mQuit.ClickedCh:
				a.isQuitting = true
				a.StopVless()
				a.shutdownWg.Wait()
				a.cleanupZombies()
				systray.Quit()
				wailsRuntime.Quit(ctx)
				return
			}
		}
	}()
}

func (a *App) onTrayExit() {
}

func (a *App) updateTrayState(connected bool) {
	if mConnect == nil || mStatus == nil {
		return
	}
	if connected {
		mConnect.SetTitle("Disconnect")
		mStatus.SetTitle("Status: Connected")
		systray.SetTooltip("Censaway: Connected")
	} else {
		mConnect.SetTitle("Connect")
		mStatus.SetTitle("Status: Disconnected")
		systray.SetTooltip("Censaway: Disconnected")
	}
}