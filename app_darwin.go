//go:build darwin

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) configureCmd(cmd *exec.Cmd) {
}

func (a *App) cleanupZombies() {
	info, err := a.loadCoreProcessInfo()
	if err != nil {
		a.log("Failed to read core process info: " + err.Error())
		return
	}
	if info == nil {
		return
	}
	if info.PID <= 0 {
		a.clearCoreProcessInfo()
		return
	}

	if info.BinPath != "" && !darwinProcessMatchesBinary(info.PID, info.BinPath) {
		a.clearCoreProcessInfo()
		return
	}

	_ = exec.Command("kill", "-9", strconv.Itoa(info.PID)).Run()
	a.clearCoreProcessInfo()
}

func darwinProcessMatchesBinary(pid int, binPath string) bool {
	out, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "command=").Output()
	if err != nil {
		return false
	}
	return strings.Contains(strings.TrimSpace(string(out)), binPath)
}

func (a *App) platformInit() error {
	return nil
}

const launchAgentTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.wails.CensawayApp</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{.ExePath}}</string>
        <string>-tray</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <false/>
</dict>
</plist>`

func (a *App) getLaunchAgentPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", "com.wails.CensawayApp.plist"), nil
}

func (a *App) EnableAutostart() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	plistPath, err := a.getLaunchAgentPath()
	if err != nil {
		return err
	}

	os.MkdirAll(filepath.Dir(plistPath), 0755)

	tmpl, err := template.New("plist").Parse(launchAgentTemplate)
	if err != nil {
		return err
	}

	var data bytes.Buffer
	if err := tmpl.Execute(&data, map[string]string{"ExePath": exe}); err != nil {
		return err
	}

	return os.WriteFile(plistPath, data.Bytes(), 0644)
}

func (a *App) DisableAutostart() error {
	plistPath, err := a.getLaunchAgentPath()
	if err != nil {
		return err
	}
	return os.Remove(plistPath)
}

func (a *App) ensurePermissions(binPath string) error {
	exec.Command("xattr", "-d", "com.apple.quarantine", binPath).Run()

	info, err := os.Stat(binPath)
	if err != nil {
		return fmt.Errorf("failed to stat core binary: %v", err)
	}

	// Hardening: never keep setuid bit on a binary under user-controlled path.
	if (info.Mode() & os.ModeSetuid) != 0 {
		cleanMode := info.Mode() &^ os.ModeSetuid
		if err := os.Chmod(binPath, cleanMode); err != nil {
			return fmt.Errorf("refusing setuid binary and failed to clear setuid bit: %v", err)
		}
		if a.ctx != nil {
			wailsRuntime.EventsEmit(a.ctx, "log", "Removed legacy setuid bit from core binary")
		}
	}

	if os.Geteuid() != 0 {
		return fmt.Errorf("administrator privileges required for TUN mode on macOS")
	}

	return nil
}
