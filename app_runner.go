package main

import (
	"bufio"
	"context"
	"errors"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) GetRunningState() bool {
	a.cmdLock.Lock()
	defer a.cmdLock.Unlock()
	return a.proxyCmd != nil
}

func (a *App) StartVless(vlessLink string) string {
	a.shutdownWg.Wait()

	if err := a.checkAndInstallCore(); err != nil {
		msg := "Core installation failed: " + err.Error()
		a.log(msg)
		return msg
	}

	a.cmdLock.Lock()
	if a.proxyCmd != nil {
		a.cmdLock.Unlock()
		return "Already running"
	}

	for _, p := range a.Profiles {
		if p.Key == vlessLink {
			a.Settings.LastProfileID = p.ID
			a.SaveSettings(a.Settings)
			break
		}
	}
	a.cmdLock.Unlock()

	for i := 0; i < 5; i++ {
		conn, err := net.DialTimeout("tcp", "127.0.0.1:9090", 200*time.Millisecond)
		if err != nil {
			break
		}
		conn.Close()
		time.Sleep(300 * time.Millisecond)
	}

	binPath, err := a.getProxyBin()
	if err != nil {
		return "Core missing"
	}

	if a.Settings.RunMode == "tun" {
		if err := a.ensurePermissions(binPath); err != nil {
			a.log("Error: Admin permissions required for TUN mode")
			return "Permission denied"
		}
	}

	workDir := a.getAppDataDir()
	configJSON, err := a.generateConfig(vlessLink)
	if err != nil {
		a.log("Config Gen Error: " + err.Error())
		return "Config error: " + err.Error()
	}

	configPath := filepath.Join(workDir, "config.json")
	os.WriteFile(configPath, []byte(configJSON), 0644)

	cmd := exec.Command(binPath, "run", "-c", configPath, "-D", workDir)
	cmd.Dir = workDir

	a.configureCmd(cmd)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "Pipe Error: " + err.Error()
	}

	if err := cmd.Start(); err != nil {
		return "Start failed: " + err.Error()
	}

	a.cmdLock.Lock()
	a.proxyCmd = cmd
	a.cmdLock.Unlock()
	a.updateTrayState(true)

	var logWg sync.WaitGroup
	logWg.Add(1)

	go func() {
		defer logWg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			text := scanner.Text()
			a.log(text)

			if strings.Contains(text, "FATAL") || strings.Contains(text, "panic") {
				wailsRuntime.EventsEmit(a.ctx, "error", "Core Error: "+text)
			}
		}

		if err := scanner.Err(); err != nil {
			if !errors.Is(err, os.ErrClosed) && !strings.Contains(err.Error(), "file already closed") {
				a.log("Log read error: " + err.Error())
			}
		}
	}()

	go func() {
		err := cmd.Wait()
		logWg.Wait()

		a.cmdLock.Lock()
		defer a.cmdLock.Unlock()

		if a.proxyCmd != nil {
			a.proxyCmd = nil

			msg := "Core process stopped unexpected"
			if err != nil {
				msg += ": " + err.Error()
			}
			a.log(msg)
			wailsRuntime.EventsEmit(a.ctx, "connection_lost", msg)
			wailsRuntime.EventsEmit(a.ctx, "connection_status", "disconnected")
			a.updateTrayState(false)

			if a.Settings.RunMode == "proxy" {
				a.setSystemProxy(false, 0)
			}
		}
	}()

	time.Sleep(1000 * time.Millisecond)

	a.cmdLock.Lock()
	if a.proxyCmd == nil {
		a.cmdLock.Unlock()
		return "Core crashed immediately (check logs)"
	}
	a.cmdLock.Unlock()

	if a.Settings.RunMode == "proxy" {
		if err := a.setSystemProxy(true, a.Settings.MixedPort); err != nil {
			a.log("Failed to set system proxy: " + err.Error())
		}
	}

	a.startStatsCollector()
	wailsRuntime.EventsEmit(a.ctx, "connection_status", "connected")
	return "Connected"
}

func (a *App) StopVless() string {
	a.stopStatsCollector()

	if a.Settings.RunMode == "proxy" {
		a.setSystemProxy(false, 0)
	}

	a.cmdLock.Lock()
	cmd := a.proxyCmd
	a.proxyCmd = nil
	a.cmdLock.Unlock()

	if cmd == nil || cmd.Process == nil {
		return "Not running"
	}

	a.shutdownWg.Add(1)
	a.updateTrayState(false)

	go func() {
		defer a.shutdownWg.Done()

		if runtime.GOOS == "windows" {
			cmd.Process.Kill()
		} else {
			cmd.Process.Signal(os.Interrupt)
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			select {
			case <-done:
			case <-time.After(1000 * time.Millisecond):
				cmd.Process.Kill()
				<-done
			}
		}
		a.log(">>> Core shutdown complete")
		wailsRuntime.EventsEmit(a.ctx, "connection_status", "disconnected")
	}()

	return "Disconnected"
}

func (a *App) startStatsCollector() {
	time.Sleep(1 * time.Second)
	url := "ws://127.0.0.1:9090/traffic?token="
	ctx, cancel := context.WithCancel(context.Background())
	a.statsCancel = cancel
	go func() {
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				c, _, err := websocket.DefaultDialer.Dial(url, nil)
				if err != nil {
					time.Sleep(1 * time.Second)
					continue
				}
				for {
					_, msg, err := c.ReadMessage()
					if err != nil {
						break
					}
					wailsRuntime.EventsEmit(a.ctx, "traffic", string(msg))
				}
				c.Close()
				time.Sleep(1 * time.Second)
			}
		}
	}()
}

func (a *App) stopStatsCollector() {
	if a.statsCancel != nil {
		a.statsCancel()
		a.statsCancel = nil
	}
}
