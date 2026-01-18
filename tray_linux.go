//go:build linux

package main

import (
	"context"
	"time"

	"github.com/energye/systray"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	mConnect *systray.MenuItem
	mStatus  *systray.MenuItem
)

func (a *App) SetupTray(ctx context.Context) {
	start, _ := systray.RunWithExternalLoop(func() { a.onTrayReady(ctx) }, a.onTrayExit)
	
	go start()
}

func (a *App) onTrayReady(ctx context.Context) {
	time.Sleep(200 * time.Millisecond)

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

	chanConnect := make(chan struct{})
	chanShow := make(chan struct{})
	chanQuit := make(chan struct{})

	mConnect.Click(func() {
		go func() { chanConnect <- struct{}{} }()
	})
	mShow.Click(func() {
		go func() { chanShow <- struct{}{} }()
	})
	mQuit.Click(func() {
		go func() { chanQuit <- struct{}{} }()
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
						wailsRuntime.WindowUnminimise(ctx)
					}
				}
			case <-chanShow:
				wailsRuntime.WindowShow(ctx)
				wailsRuntime.WindowUnminimise(ctx)
				wailsRuntime.WindowSetAlwaysOnTop(ctx, true)
				time.Sleep(100 * time.Millisecond)
				wailsRuntime.WindowSetAlwaysOnTop(ctx, false)
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