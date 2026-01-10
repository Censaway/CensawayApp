package main

import (
	"encoding/json"
	"os"
)

func (a *App) LoadSettings() Settings {
	data, err := os.ReadFile(a.getSettingsPath())
	if err == nil {
		json.Unmarshal(data, &a.Settings)
	}
	if a.Settings.RoutingMode == "" {
		a.Settings.RoutingMode = "smart"
	}
	if a.Settings.RunMode == "" {
		a.Settings.RunMode = "tun"
	}
	if a.Settings.MixedPort == 0 {
		a.Settings.MixedPort = 2080
	}
	if len(a.Settings.RuDomains) == 0 {
		a.Settings.RuDomains = defaultRuDomains
	}
	return a.Settings
}

func (a *App) SaveSettings(s Settings) string {
	a.Settings = s
	data, err := json.MarshalIndent(a.Settings, "", "  ")
	if err != nil {
		return "Error"
	}
	os.WriteFile(a.getSettingsPath(), data, 0644)

	if s.AutoConnect {
		a.EnableAutostart()
	} else {
		a.DisableAutostart()
	}

	return "Saved"
}

func (a *App) GetSettings() Settings { return a.Settings }
