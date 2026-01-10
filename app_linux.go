//go:build linux

package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

func (a *App) configureCmd(cmd *exec.Cmd) {
}

func (a *App) EnableAutostart() error {
	return a.installDesktopFile(true)
}

func (a *App) DisableAutostart() error {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".config", "autostart", "censaway.desktop")
	return os.Remove(path)
}

func (a *App) platformInit() error {
	iconPath, err := a.installIcon()
	if err != nil {
		return err
	}
	return a.installDesktopFile(false, iconPath)
}

func (a *App) installIcon() (string, error) {
	home, _ := os.UserHomeDir()
	iconDir := filepath.Join(home, ".local", "share", "icons")
	if err := os.MkdirAll(iconDir, 0755); err != nil {
		return "", err
	}

	iconPath := filepath.Join(iconDir, "censaway.png")

	if info, err := os.Stat(iconPath); err != nil || info.Size() == 0 {
		if len(a.Icon) > 0 {
			if err := os.WriteFile(iconPath, a.Icon, 0644); err != nil {
				return "", err
			}
		}
	}

	return iconPath, nil
}

func (a *App) installDesktopFile(autostart bool, explicitIconPath ...string) error {
	home, _ := os.UserHomeDir()
	var targetDir string
	if autostart {
		targetDir = filepath.Join(home, ".config", "autostart")
	} else {
		targetDir = filepath.Join(home, ".local", "share", "applications")
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	exe, _ := os.Executable()
	execCmd := exe
	if autostart {
		execCmd = exe + " -tray"
	}

	iconValue := "censaway"
	if len(explicitIconPath) > 0 && explicitIconPath[0] != "" {
		iconValue = explicitIconPath[0]
	}

	content := `[Desktop Entry]
Type=Application
Name=Censaway
Exec=` + execCmd + `
Icon=` + iconValue + `
Comment=Censaway VPN Client
Terminal=false
Categories=Network;
StartupWMClass=CensawayApp
`
	return os.WriteFile(filepath.Join(targetDir, "censaway.desktop"), []byte(content), 0644)
}

func (a *App) cleanupZombies() {
	exec.Command("pkill", "sing-box").Run()
}
