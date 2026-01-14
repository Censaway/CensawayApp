package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) LoadProfiles() []Profile {
	data, err := os.ReadFile(a.getProfilesPath())
	if err != nil {
		return []Profile{}
	}
	json.Unmarshal(data, &a.Profiles)
	if a.Profiles == nil {
		a.Profiles = []Profile{}
	}
	return a.Profiles
}

func (a *App) SaveProfiles() error {
	data, err := json.MarshalIndent(a.Profiles, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(a.getProfilesPath(), data, 0644)
}

func (a *App) AddProfile(vlessLink string) string {
	if !strings.HasPrefix(vlessLink, "vless://") {
		return "Invalid VLESS link"
	}
	u, err := url.Parse(vlessLink)
	if err != nil {
		return "Parse error"
	}
	name := u.Fragment
	if name == "" {
		name = u.Hostname()
	}
	name, _ = url.QueryUnescape(name)
	a.Profiles = append(a.Profiles, Profile{ID: uuid.New().String(), Name: name, Key: vlessLink, CreatedAt: time.Now().Unix()})
	a.SaveProfiles()
	return "OK"
}

func (a *App) DeleteProfile(id string) []Profile {
	newP := []Profile{}
	for _, p := range a.Profiles {
		if p.ID != id {
			newP = append(newP, p)
		}
	}
	a.Profiles = newP
	a.SaveProfiles()
	return a.Profiles
}

func (a *App) GetProfiles() []Profile { return a.LoadProfiles() }

func (a *App) ImportSubscription(subUrl string) string {
	resp, err := http.Get(subUrl)
	if err != nil {
		return "Error: " + err.Error()
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	decoded, err := base64.StdEncoding.DecodeString(string(body))
	if err != nil {
		decoded, _ = base64.URLEncoding.DecodeString(string(body))
	}
	content := string(body)
	if len(decoded) > 0 {
		content = string(decoded)
	}
	count := 0
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "vless://") {
			a.AddProfile(strings.TrimSpace(line))
			count++
		}
	}
	return fmt.Sprintf("Imported %d profiles", count)
}

func (a *App) TcpPing(profileID string) int {
	var targetKey string
	for _, p := range a.Profiles {
		if p.ID == profileID {
			targetKey = p.Key
			break
		}
	}

	if targetKey == "" {
		wailsRuntime.EventsEmit(a.ctx, "log", "Ping: Profile not found")
		return -1
	}

	u, err := url.Parse(targetKey)
	if err != nil {
		wailsRuntime.EventsEmit(a.ctx, "log", "Ping: URL parse error: "+err.Error())
		return -1
	}

	port := u.Port()
	if port == "" {
		port = "443"
	}

	host := u.Hostname()
	target := net.JoinHostPort(host, port)

	var conn net.Conn
	var dialErr error
	start := time.Now()

	for i := 0; i < 2; i++ {
		conn, dialErr = net.DialTimeout("tcp", target, 3*time.Second)
		if dialErr == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	if dialErr != nil {
		wailsRuntime.EventsEmit(a.ctx, "log", fmt.Sprintf("Ping: Connection failed to %s: %v", target, dialErr))
		return -1
	}
	conn.Close()

	return int(time.Since(start).Milliseconds())
}

func (a *App) UpdateProfile(id string, name string, key string) string {
	if !strings.HasPrefix(key, "vless://") {
		return "Invalid VLESS key"
	}

	found := false
	for i, p := range a.Profiles {
		if p.ID == id {
			a.Profiles[i].Name = name
			a.Profiles[i].Key = key
			found = true
			break
		}
	}

	if !found {
		return "Profile not found"
	}

	if err := a.SaveProfiles(); err != nil {
		return "Save failed: " + err.Error()
	}
	return "OK"
}
