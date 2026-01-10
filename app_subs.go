package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (a *App) getSubscriptionsPath() string {
	return filepath.Join(a.getAppDataDir(), "subscriptions.json")
}

func (a *App) LoadSubscriptions() []Subscription {
	data, err := os.ReadFile(a.getSubscriptionsPath())
	if err != nil {
		return []Subscription{}
	}
	json.Unmarshal(data, &a.Subscriptions)
	if a.Subscriptions == nil {
		a.Subscriptions = []Subscription{}
	}
	return a.Subscriptions
}

func (a *App) SaveSubscriptions() error {
	data, err := json.MarshalIndent(a.Subscriptions, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(a.getSubscriptionsPath(), data, 0644)
}

func (a *App) CreateSubscription(subUrl string) string {
	_, err := url.Parse(subUrl)
	if err != nil {
		return "Invalid URL"
	}

	subID := uuid.New().String()
	name := "New Subscription"

	newSub := Subscription{
		ID:        subID,
		Name:      name,
		Url:       subUrl,
		UpdatedAt: 0,
	}

	a.Subscriptions = append(a.Subscriptions, newSub)
	a.SaveSubscriptions()

	return a.UpdateSubscription(subID)
}

func (a *App) DeleteSubscription(subID string) {
	newSubs := []Subscription{}
	for _, s := range a.Subscriptions {
		if s.ID != subID {
			newSubs = append(newSubs, s)
		}
	}
	a.Subscriptions = newSubs
	a.SaveSubscriptions()

	newProfs := []Profile{}
	for _, p := range a.Profiles {
		if p.SubscriptionID != subID {
			newProfs = append(newProfs, p)
		}
	}
	a.Profiles = newProfs
	a.SaveProfiles()
}

func (a *App) UpdateSubscription(subID string) string {
	var targetSub *Subscription
	for i := range a.Subscriptions {
		if a.Subscriptions[i].ID == subID {
			targetSub = &a.Subscriptions[i]
			break
		}
	}
	if targetSub == nil {
		return "Subscription not found"
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(targetSub.Url)
	if err != nil {
		return "Download failed: " + err.Error()
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	content := string(body)

	decoded, err := base64.StdEncoding.DecodeString(content)
	if err == nil {
		content = string(decoded)
	} else {
		decoded, err = base64.URLEncoding.DecodeString(content)
		if err == nil {
			content = string(decoded)
		}
	}

	var newLinks []string
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "vless://") {
			newLinks = append(newLinks, line)
		}
	}

	if len(newLinks) == 0 {
		return "No valid links found"
	}

	targetSub.UpdatedAt = time.Now().Unix()

	tempProfiles := []Profile{}
	for _, p := range a.Profiles {
		if p.SubscriptionID != subID {
			tempProfiles = append(tempProfiles, p)
		}
	}
	a.Profiles = tempProfiles

	for _, link := range newLinks {
		u, _ := url.Parse(link)
		name := u.Fragment
		if name == "" {
			name = u.Hostname()
		}
		name, _ = url.QueryUnescape(name)

		a.Profiles = append(a.Profiles, Profile{
			ID:             uuid.New().String(),
			Name:           name,
			Key:            link,
			SubscriptionID: subID,
			CreatedAt:      time.Now().Unix(),
		})
	}

	a.SaveSubscriptions()
	a.SaveProfiles()

	return fmt.Sprintf("Updated: %d profiles", len(newLinks))
}

func (a *App) GetSubscriptions() []Subscription {
	return a.LoadSubscriptions()
}
