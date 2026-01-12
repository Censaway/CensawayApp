package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type AppUpdateInfo struct {
	Available   bool   `json:"available"`
	Version     string `json:"version"`
	CurrentVer  string `json:"current_ver"`
	ReleaseUrl  string `json:"release_url"`
	DownloadUrl string `json:"download_url"`
	Body        string `json:"body"`
}

const UpdateRepo = "Censaway/CensawayApp"

func (a *App) CheckAppUpdate() AppUpdateInfo {
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", UpdateRepo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return AppUpdateInfo{Available: false}
	}

	req.Header.Set("User-Agent", "CensawayApp")

	resp, err := client.Do(req)
	if err != nil {
		return AppUpdateInfo{Available: false}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return AppUpdateInfo{Available: false}
	}

	var release struct {
		TagName string `json:"tag_name"`
		HtmlUrl string `json:"html_url"`
		Body    string `json:"body"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return AppUpdateInfo{Available: false}
	}

	if isNewerVersion(release.TagName, AppVersion) {
		return AppUpdateInfo{
			Available:   true,
			Version:     release.TagName,
			CurrentVer:  AppVersion,
			ReleaseUrl:  release.HtmlUrl,
			Body:        release.Body,
		}
	}

	return AppUpdateInfo{Available: false, CurrentVer: AppVersion}
}

func (a *App) OpenUrl(url string) {
	wailsRuntime.BrowserOpenURL(a.ctx, url)
}

func isNewerVersion(remoteVer, currentVer string) bool {
	parse := func(v string) []int {
		v = strings.TrimPrefix(v, "v")
		parts := strings.Split(v, ".")
		res := make([]int, 0, len(parts))
		for _, part := range parts {
			if idx := strings.Index(part, "-"); idx != -1 {
				part = part[:idx]
			}
			val, err := strconv.Atoi(part)
			if err == nil {
				res = append(res, val)
			}
		}
		return res
	}

	remote := parse(remoteVer)
	current := parse(currentVer)

	maxLen := len(remote)
	if len(current) > maxLen {
		maxLen = len(current)
	}

	for i := 0; i < maxLen; i++ {
		rVal := 0
		if i < len(remote) {
			rVal = remote[i]
		}

		cVal := 0
		if i < len(current) {
			cVal = current[i]
		}

		if rVal > cVal {
			return true
		}
		if rVal < cVal {
			return false
		}
	}

	return false
}
