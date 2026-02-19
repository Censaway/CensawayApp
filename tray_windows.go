//go:build windows

package main

import (
	"context"
	_ "embed"

	"github.com/energye/systray"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed build/windows/icon.ico
var trayIconIco []byte

var (
	mConnect *systray.MenuItem
	mStatus  *systray.MenuItem
)

func (a *App) SetupTray(ctx context.Context) {
	start, _ := systray.RunWithExternalLoop(func() { a.onTrayReady(ctx) }, a.onTrayExit)
	go start()
}

func (a *App) onTrayReady(ctx context.Context) {
	if len(trayIconIco) > 0 {
		systray.SetIcon(trayIconIco)
	} else {
		systray.SetIcon(a.Icon)
	}

	systray.SetTitle("Censaway")
	systray.SetTooltip("CensawayApp VPN")

	mStatus = systray.AddMenuItem("Status: Disconnected", "Current VPN status")
	mStatus.Disable()

	systray.AddSeparator()

	mConnect = systray.AddMenuItem("Connect", "Toggle VPN connection")
	mShow := systray.AddMenuItem("Show Window", "Show main window")

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Exit app")

	chanConnect := make(chan struct{}, 1)
	chanShow := make(chan struct{}, 1)
	chanQuit := make(chan struct{}, 1)

	mConnect.Click(func() {
		select {
		case chanConnect <- struct{}{}:
		default:
		}
	})
	mShow.Click(func() {
		select {
		case chanShow <- struct{}{}:
		default:
		}
	})
	mQuit.Click(func() {
		select {
		case chanQuit <- struct{}{}:
		default:
		}
	})

	go func() {
		for {
			select {
			case <-chanConnect:
				if a.GetRunningState() {
					a.StopVless()
				} else {
					if a.Settings.LastProfileID != "" {
						a.connectLastProfile()
					} else {
						wailsRuntime.WindowShow(ctx)
					}
				}
			case <-chanShow:
				wailsRuntime.WindowShow(ctx)
				wailsRuntime.WindowUnminimise(ctx)
			case <-chanQuit:
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
	if mConnect == nil {
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
