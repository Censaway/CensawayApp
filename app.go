package main

import (
	"context"
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed dll/wintun.dll
var wintunDll []byte

type Subscription struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Url       string `json:"url"`
	UpdatedAt int64  `json:"updated_at"`
}

type Profile struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Key            string `json:"key"`
	SubscriptionID string `json:"subscription_id"`
	CreatedAt      int64  `json:"created_at"`
}

type Settings struct {
	RoutingMode   string     `json:"routing_mode"`
	RunMode       string     `json:"run_mode"`
	MixedPort     int        `json:"mixed_port"`
	UserRules     []UserRule `json:"user_rules"`
	RuDomains     []string   `json:"ru_domains"`
	AutoConnect   bool       `json:"auto_connect"`
	LastProfileID string     `json:"last_profile_id"`
}

type UserRule struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Value    string `json:"value"`
	Outbound string `json:"outbound"`
}

type App struct {
	ctx           context.Context
	proxyCmd      *exec.Cmd
	cmdLock       sync.Mutex
	shutdownWg    sync.WaitGroup
	Profiles      []Profile
	Subscriptions []Subscription
	Settings      Settings
	statsCancel   context.CancelFunc
	isQuitting    bool
	Icon          []byte

	logBuffer []string
	logLock   sync.Mutex
}

var defaultRuDomains = []string{
	".ru", ".rf", ".xn--p1ai",
}

func NewApp() *App {
	return &App{
		Profiles:      []Profile{},
		Subscriptions: []Subscription{},
		Settings: Settings{
			RoutingMode: "smart",
			RunMode:     "tun",
			MixedPort:   2080,
			UserRules:   []UserRule{},
			RuDomains:   defaultRuDomains,
		},
		isQuitting: false,
		logBuffer:  make([]string, 0, 100),
	}
}

func (a *App) log(msg string) {
	a.logLock.Lock()
	if len(a.logBuffer) >= 100 {
		a.logBuffer = a.logBuffer[1:]
	}
	a.logBuffer = append(a.logBuffer, msg)
	a.logLock.Unlock()

	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, "log", msg)
	}
}

func (a *App) GetLogs() []string {
	a.logLock.Lock()
	defer a.logLock.Unlock()
	logs := make([]string, len(a.logBuffer))
	copy(logs, a.logBuffer)
	return logs
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	_ = a.getAppDataDir()

	a.cleanupZombies()
	a.ensureProxyDisabled()

	a.LoadSettings()
	a.LoadProfiles()
	a.LoadSubscriptions()
	
	a.platformInit()

	go func() {
		if err := a.ensureWintun(); err != nil {
			a.log("Failed to extract wintun.dll: " + err.Error())
		}

		err := a.checkAndInstallCore()
		if err != nil {
			a.log("Failed to install core: " + err.Error())
			wailsRuntime.EventsEmit(a.ctx, "error", "Core Install Error")
		} else {
			if a.Settings.AutoConnect && a.Settings.LastProfileID != "" {
				a.connectLastProfile()
			}
		}
	}()
}

func (a *App) connectLastProfile() {
	var targetLink string
	for _, p := range a.Profiles {
		if p.ID == a.Settings.LastProfileID {
			targetLink = p.Key
			break
		}
	}
	if targetLink != "" {
		a.log("Auto-connecting...")
		time.Sleep(1 * time.Second)
		res := a.StartVless(targetLink)
		if res != "Connected" {
			a.log("Auto-connect ERROR: " + res)
		}
	}
}

func (a *App) getAppDataDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config")
	}
	appDir := filepath.Join(configDir, "CensawayApp")
	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		os.MkdirAll(appDir, 0755)
	}
	return appDir
}

func (a *App) getProfilesPath() string { return filepath.Join(a.getAppDataDir(), "profiles.json") }
func (a *App) getSettingsPath() string { return filepath.Join(a.getAppDataDir(), "settings.json") }
func (a *App) getGeoIpPath() string    { return filepath.Join(a.getAppDataDir(), "geoip.dat") }
func (a *App) getSrsPath() string      { return filepath.Join(a.getAppDataDir(), "geoip-ru.srs") }