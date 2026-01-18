//go:build darwin

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
	
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) configureCmd(cmd *exec.Cmd) {
}

func (a *App) cleanupZombies() {
	exec.Command("pkill", "sing-box").Run()
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
	if err == nil {
		if (info.Mode() & os.ModeSetuid) != 0 {
			return nil
		}
	}

	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, "log", "Requesting admin rights (osascript)...")
	}

	cmdStr := fmt.Sprintf("chown root:admin \\\"%s\\\" && chmod +s \\\"%s\\\"", binPath, binPath)
	script := fmt.Sprintf("do shell script \"%s\" with administrator privileges", cmdStr)

	err = exec.Command("osascript", "-e", script).Run()
	if err != nil {
		return fmt.Errorf("failed to set SUID: %v", err)
	}
	return nil
}