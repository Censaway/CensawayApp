package main

import (
	"encoding/json"
	"os"
	"reflect"
	"runtime"
)

const settingsFileMode os.FileMode = 0600
const defaultMixedPort = 2080

func defaultRunMode() string {
	if runtime.GOOS == "darwin" {
		return "proxy"
	}
	return "tun"
}

func normalizeSettingsDefaults(s Settings) Settings {
	if s.Language != "en" && s.Language != "ru" {
		s.Language = "en"
	}
	if s.RoutingMode != "smart" && s.RoutingMode != "global" {
		s.RoutingMode = "smart"
	}
	if s.RunMode != "tun" && s.RunMode != "proxy" {
		s.RunMode = defaultRunMode()
	}
	if s.MixedPort < 1 || s.MixedPort > 65535 {
		s.MixedPort = defaultMixedPort
	}
	if len(s.RuDomains) == 0 {
		s.RuDomains = append([]string{}, defaultRuDomains...)
	}
	if s.UserRules == nil {
		s.UserRules = []UserRule{}
	}
	return s
}

func (a *App) LoadSettings() Settings {
	data, err := os.ReadFile(a.getSettingsPath())

	if os.IsNotExist(err) {
		a.Settings = Settings{
			Language:    "en",
			RoutingMode: "smart",
			RunMode:     defaultRunMode(),
			MixedPort:   defaultMixedPort,
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

	normalized := normalizeSettingsDefaults(a.Settings)
	// Existing macOS users may have persisted TUN from older defaults.
	// Keep app usable for non-elevated launches by migrating to proxy mode.
	if runtime.GOOS == "darwin" && os.Geteuid() != 0 && normalized.RunMode == "tun" {
		normalized.RunMode = "proxy"
	}
	updated := !reflect.DeepEqual(a.Settings, normalized)
	a.Settings = normalized
	if updated {
		a.SaveSettings(a.Settings)
	}

	return a.Settings
}

func (a *App) SaveSettings(s Settings) string {
	if s.MixedPort < 1 || s.MixedPort > 65535 {
		return "Invalid mixed port: must be 1-65535"
	}

	normalized := normalizeSettingsDefaults(s)
	a.Settings = normalized
	data, err := json.MarshalIndent(a.Settings, "", "  ")
	if err != nil {
		return "Error"
	}

	err = os.WriteFile(a.getSettingsPath(), data, settingsFileMode)
	if err != nil {
		return "Write Failed: " + err.Error()
	}

	if normalized.AutoConnect {
		if err := a.EnableAutostart(); err != nil {
			a.log("Autostart enable failed: " + err.Error())
			return "Autostart enable failed: " + err.Error()
		}
	} else {
		if err := a.DisableAutostart(); err != nil && !os.IsNotExist(err) {
			a.log("Autostart disable failed: " + err.Error())
			return "Autostart disable failed: " + err.Error()
		}
	}

	return "Saved"
}

func (a *App) GetSettings() Settings { return a.Settings }
