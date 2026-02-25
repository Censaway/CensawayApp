//go:build linux

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
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

	startupWMClass := "censawayapp"

	buildContent := func(noDisplay bool) string {
		content := `[Desktop Entry]
Type=Application
Name=Censaway
Exec=` + execCmd + `
Icon=` + iconValue + `
Comment=Censaway VPN Client
Terminal=false
Categories=Network;
StartupWMClass=` + startupWMClass + `
`
		if noDisplay {
			content += "NoDisplay=true\n"
		}
		return content
	}

	primaryFile := filepath.Join(targetDir, "censaway.desktop")
	if err := os.WriteFile(primaryFile, []byte(buildContent(false)), 0644); err != nil {
		return err
	}

	// waybar/wlroots taskbars often resolve icons by desktop-id/app_id.
	// In `wails dev`, app_id is typically derived from the dev binary name.
	if autostart {
		return nil
	}

	aliases := desktopAliasesForExe(exe)
	for _, alias := range aliases {
		if alias == "censaway.desktop" {
			continue
		}
		if err := os.WriteFile(filepath.Join(targetDir, alias), []byte(buildContent(true)), 0644); err != nil {
			return err
		}
	}

	return nil
}

func desktopAliasesForExe(exe string) []string {
	base := filepath.Base(exe)
	candidates := []string{
		"censaway.desktop",
		"censawayapp.desktop",
		"CensawayApp.desktop",
	}
	if base != "" {
		candidates = append(candidates, base+".desktop")
		lowerBase := strings.ToLower(base)
		if lowerBase != base {
			candidates = append(candidates, lowerBase+".desktop")
		}
	}

	seen := make(map[string]struct{}, len(candidates))
	aliases := make([]string, 0, len(candidates))
	for _, name := range candidates {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		aliases = append(aliases, name)
	}
	return aliases
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

	if info.BinPath != "" && !linuxProcessMatchesBinary(info.PID, info.BinPath) {
		a.clearCoreProcessInfo()
		return
	}

	_ = exec.Command("kill", "-9", strconv.Itoa(info.PID)).Run()
	a.clearCoreProcessInfo()
}

func linuxProcessMatchesBinary(pid int, binPath string) bool {
	out, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "args=").Output()
	if err != nil {
		return false
	}
	return strings.Contains(strings.TrimSpace(string(out)), binPath)
}

func (a *App) ensurePermissions(binPath string) error {
	checkCmd := exec.Command("getcap", binPath)
	out, _ := checkCmd.CombinedOutput()
	if strings.Contains(string(out), "cap_net_admin") {
		return nil
	}

	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, "log", "Requesting admin rights (pkexec)...")
	}

	err := exec.Command("pkexec", "setcap", "cap_net_admin,cap_net_bind_service=+ep", binPath).Run()
	if err != nil {
		return exec.Command("sudo", "setcap", "cap_net_admin,cap_net_bind_service=+ep", binPath).Run()
	}
	return nil
}
