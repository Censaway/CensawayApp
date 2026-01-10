package main

import (
	"fmt"
	"net/http"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/net/proxy"
)

func (a *App) UrlTest(profileID string) int {
	a.cmdLock.Lock()
	isRunning := a.proxyCmd != nil
	currentSettings := a.Settings
	a.cmdLock.Unlock()

	if isRunning {
		proxyAddr := fmt.Sprintf("127.0.0.1:%d", currentSettings.MixedPort)

		dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
		if err != nil {
			wailsRuntime.EventsEmit(a.ctx, "log", fmt.Sprintf("Latency: SOCKS5 init error: %v", err))
			return -1
		}

		httpClient := &http.Client{
			Transport: &http.Transport{Dial: dialer.Dial},
			Timeout:   5 * time.Second,
		}

		start := time.Now()
		resp, err := httpClient.Head("http://www.gstatic.com/generate_204")
		if err != nil {
			wailsRuntime.EventsEmit(a.ctx, "log", fmt.Sprintf("Latency: Request failed: %v", err))
			return -1
		}
		resp.Body.Close()

		if resp.StatusCode != 204 && resp.StatusCode != 200 {
			wailsRuntime.EventsEmit(a.ctx, "log", fmt.Sprintf("Latency: Bad status code: %d", resp.StatusCode))
			return -1
		}

		return int(time.Since(start).Milliseconds())
	}

	return a.TcpPing(profileID)
}
