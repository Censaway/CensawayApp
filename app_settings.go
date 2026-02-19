package main

import (
	"encoding/json"
	"os"
)

func (a *App) LoadSettings() Settings {
	data, err := os.ReadFile(a.getSettingsPath())
	
	if os.IsNotExist(err) {
		a.Settings = Settings{
			Language:    "en",
			RoutingMode: "smart",
			RunMode:     "tun",
			MixedPort:   2080,
			RuDomains:   defaultRuDomains,
			UserRules:   []UserRule{},
		}
		a.SaveSettings(a.Settings)
		return a.Settings
	}

	if err != nil {
		a.log("Error reading settings file: " + err.Error())
		return a.Settings
	}

	err = json.Unmarshal(data, &a.Settings)
	if err != nil {
		a.log("Error parsing settings JSON: " + err.Error())
		return a.Settings
	}

	updated := false
	if a.Settings.Language == "" {
		a.Settings.Language = "en"
		updated = true
	}
	if a.Settings.RoutingMode == "" {
		a.Settings.RoutingMode = "smart"
		updated = true
	}
	if a.Settings.RunMode == "" {
		a.Settings.RunMode = "tun"
		updated = true
	}
	if a.Settings.MixedPort == 0 {
		a.Settings.MixedPort = 2080
		updated = true
	}
	if len(a.Settings.RuDomains) == 0 {
		a.Settings.RuDomains = defaultRuDomains
		updated = true
	}
	if a.Settings.UserRules == nil {
		a.Settings.UserRules = []UserRule{}
		updated = true
	}

	if updated {
		a.SaveSettings(a.Settings)
	}

	return a.Settings
}

func (a *App) SaveSettings(s Settings) string {
	a.Settings = s
	data, err := json.MarshalIndent(a.Settings, "", "  ")
	if err != nil {
		return "Error"
	}
	
	err = os.WriteFile(a.getSettingsPath(), data, 0666)
	if err != nil {
		return "Write Failed: " + err.Error()
	}
	
	os.Chmod(a.getSettingsPath(), 0666)

	if s.AutoConnect {
		a.EnableAutostart()
	} else {
		a.DisableAutostart()
	}

	return "Saved"
}

func (a *App) GetSettings() Settings { return a.Settings }
