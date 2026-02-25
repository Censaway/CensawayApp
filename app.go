package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed dll/wintun.dll
var wintunDll []byte

const AppVersion = "v1.3.3"

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

type MixedProfile struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	RelayID string `json:"relay_id"`
	ExitID  string `json:"exit_id"`
}

type Settings struct {
	Language      string     `json:"language"`
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

type PortalServer struct {
	Tag  string `json:"tag"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type portalKeyInfo struct {
	UUID        string
	Host        string
	PanelSecret string
}

type portalAuthResponse struct {
	Email             string `json:"email"`
	PreferredOutbound string `json:"preferred_outbound"`
	Status            string `json:"status"`
	UUID              string `json:"uuid"`
}

type portalRoutingResponse struct {
	Status string `json:"status"`
	Tag    string `json:"tag"`
}

type CoreProcessInfo struct {
	PID     int    `json:"pid"`
	BinPath string `json:"bin_path"`
}

type App struct {
	ctx           context.Context
	proxyCmd      *exec.Cmd
	cmdLock       sync.Mutex
	shutdownWg    sync.WaitGroup
	Profiles      []Profile
	Subscriptions []Subscription
	MixedProfiles []MixedProfile
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
		MixedProfiles: []MixedProfile{},
		Settings: Settings{
			Language:    "en",
			RoutingMode: "smart",
			RunMode:     defaultRunMode(),
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
	a.LoadMixedProfiles()

	if err := a.platformInit(); err != nil {
		a.log("Platform init failed: " + err.Error())
	}

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

	a.StartUpdateTicker()
}

func (a *App) connectLastProfile() {
	var targetLink string

	if strings.HasPrefix(a.Settings.LastProfileID, "mixed://") {
		targetLink = a.Settings.LastProfileID
	} else {
		for _, p := range a.Profiles {
			if p.ID == a.Settings.LastProfileID {
				targetLink = p.Key
				break
			}
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

func (a *App) GetProxyIP() string {
	proxyUrl, _ := url.Parse(fmt.Sprintf("socks5://127.0.0.1:%d", a.Settings.MixedPort))
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get("http://checkip.amazonaws.com")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return strings.TrimSpace(string(body))
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
func (a *App) getMixedProfilesPath() string {
	return filepath.Join(a.getAppDataDir(), "mixed_profiles.json")
}
func (a *App) GetAppVersion() string { return AppVersion }
func (a *App) getCoreProcessInfoPath() string {
	return filepath.Join(a.getAppDataDir(), "sing-box.process.json")
}

func (a *App) getHttpClientForPortal() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Proxy: func(_ *http.Request) (*url.URL, error) {
				return nil, nil
			},
		},
	}
}

func (a *App) saveCoreProcessInfo(pid int, binPath string) {
	info := CoreProcessInfo{
		PID:     pid,
		BinPath: binPath,
	}

	data, err := json.Marshal(info)
	if err != nil {
		a.log("Failed to serialize core process info: " + err.Error())
		return
	}

	if err := os.WriteFile(a.getCoreProcessInfoPath(), data, 0600); err != nil {
		a.log("Failed to write core process info: " + err.Error())
	}
}

func (a *App) loadCoreProcessInfo() (*CoreProcessInfo, error) {
	data, err := os.ReadFile(a.getCoreProcessInfoPath())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var info CoreProcessInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

func (a *App) clearCoreProcessInfo() {
	path := a.getCoreProcessInfoPath()
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		a.log("Failed to remove core process info: " + err.Error())
	}
}

func (a *App) invalidateLastProfileIfMissing() {
	lastProfileID := strings.TrimSpace(a.Settings.LastProfileID)
	if lastProfileID == "" || strings.HasPrefix(lastProfileID, "mixed://") {
		return
	}

	for _, p := range a.Profiles {
		if p.ID == lastProfileID {
			return
		}
	}

	a.Settings.LastProfileID = ""
	if result := a.SaveSettings(a.Settings); result != "Saved" {
		a.log("Failed to clear stale last profile id: " + result)
	}
}

func (a *App) findProfileByID(profileID string) *Profile {
	for i := range a.Profiles {
		if a.Profiles[i].ID == profileID {
			return &a.Profiles[i]
		}
	}
	return nil
}

func parsePortalKeyInfo(vlessKey string) (*portalKeyInfo, error) {
	u, err := url.Parse(vlessKey)
	if err != nil {
		return nil, fmt.Errorf("invalid link")
	}

	info := &portalKeyInfo{
		UUID: strings.TrimSpace(u.User.Username()),
		Host: strings.TrimSpace(u.Hostname()),
	}
	if info.UUID == "" {
		return nil, fmt.Errorf("missing uuid")
	}
	if info.Host == "" {
		return nil, fmt.Errorf("missing host")
	}

	q := u.Query()
	panelSecret := strings.TrimSpace(q.Get("panel_secret"))
	if panelSecret == "" {
		panelSecret = strings.TrimSpace(q.Get("ps"))
	}
	if panelSecret == "" {
		panelSecret = strings.TrimSpace(q.Get("portal_secret"))
	}
	panelSecret = strings.Trim(panelSecret, "/")
	if panelSecret == "" {
		return nil, fmt.Errorf("missing panel secret")
	}

	info.PanelSecret = panelSecret
	return info, nil
}

func portalURL(host, panelSecret, endpoint string) string {
	escapedSecret := url.PathEscape(strings.Trim(panelSecret, "/"))
	escapedEndpoint := strings.TrimPrefix(endpoint, "/")
	return fmt.Sprintf("https://%s/%s/api/user-portal/%s", host, escapedSecret, escapedEndpoint)
}

func (a *App) postPortalJSON(client *http.Client, endpoint string, payload any) ([]byte, int, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(jsonData))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, resp.StatusCode, readErr
	}

	return body, resp.StatusCode, nil
}

func decodePortalServersResponse(body []byte) ([]PortalServer, error) {
	var servers []PortalServer
	if err := json.Unmarshal(body, &servers); err == nil {
		return servers, nil
	}

	var envelope struct {
		Status  string         `json:"status"`
		Servers []PortalServer `json:"servers"`
		Data    []PortalServer `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, err
	}
	if len(envelope.Servers) > 0 {
		return envelope.Servers, nil
	}
	if len(envelope.Data) > 0 {
		return envelope.Data, nil
	}
	return []PortalServer{}, nil
}

func (a *App) GetPortalServers(profileID string) []PortalServer {
	profile := a.findProfileByID(profileID)
	if profile == nil {
		return []PortalServer{}
	}

	info, err := parsePortalKeyInfo(profile.Key)
	if err != nil {
		a.log("Portal config error: " + err.Error())
		return []PortalServer{}
	}

	client := a.getHttpClientForPortal()
	payload := map[string]string{"uuid": info.UUID}

	authEndpoint := portalURL(info.Host, info.PanelSecret, "auth")
	authBody, authStatus, err := a.postPortalJSON(client, authEndpoint, payload)
	if err != nil {
		a.log("Portal auth request failed: " + err.Error())
		return []PortalServer{}
	}
	if authStatus != http.StatusOK {
		a.log(fmt.Sprintf("Portal auth failed: %d %s", authStatus, strings.TrimSpace(string(authBody))))
		return []PortalServer{}
	}

	var authResp portalAuthResponse
	if err := json.Unmarshal(authBody, &authResp); err != nil {
		a.log("Portal auth parse error: " + err.Error())
		return []PortalServer{}
	}
	if authResp.Status != "" && !strings.EqualFold(authResp.Status, "ok") {
		a.log("Portal auth rejected: " + authResp.Status)
		return []PortalServer{}
	}

	serversEndpoint := portalURL(info.Host, info.PanelSecret, "servers")
	serversBody, serversStatus, err := a.postPortalJSON(client, serversEndpoint, payload)
	if err != nil {
		a.log("Portal servers request failed: " + err.Error())
		return []PortalServer{}
	}
	if serversStatus != http.StatusOK {
		a.log(fmt.Sprintf("Portal servers failed: %d %s", serversStatus, strings.TrimSpace(string(serversBody))))
		return []PortalServer{}
	}

	servers, err := decodePortalServersResponse(serversBody)
	if err != nil {
		a.log("Portal servers parse error: " + err.Error())
		return []PortalServer{}
	}
	return servers
}

func (a *App) SetPortalRouting(profileID string, tag string) string {
	profile := a.findProfileByID(profileID)
	if profile == nil {
		return "Profile not found"
	}

	info, err := parsePortalKeyInfo(profile.Key)
	if err != nil {
		return "Portal config error: " + err.Error()
	}

	tag = strings.TrimSpace(tag)
	if tag == "" {
		return "Tag is required"
	}

	client := a.getHttpClientForPortal()
	endpoint := portalURL(info.Host, info.PanelSecret, "routing")
	body, statusCode, err := a.postPortalJSON(client, endpoint, map[string]string{
		"uuid": info.UUID,
		"tag":  tag,
	})
	if err != nil {
		return "Error: " + err.Error()
	}
	if statusCode != http.StatusOK {
		return fmt.Sprintf("Failed: %d %s", statusCode, strings.TrimSpace(string(body)))
	}

	var routingResp portalRoutingResponse
	if err := json.Unmarshal(body, &routingResp); err == nil {
		state := strings.ToLower(strings.TrimSpace(routingResp.Status))
		if state == "" || state == "ok" || state == "updated" {
			return "OK"
		}
		return "Failed: " + routingResp.Status
	}

	return "OK"
}
