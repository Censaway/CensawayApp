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
	return os.WriteFile(a.getSubscriptionsPath(), data, 0600)
}

func (a *App) CreateSubscription(subUrl string) string {
	parsedURL, err := url.Parse(subUrl)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
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
	if err := a.SaveSubscriptions(); err != nil {
		return "Save failed: " + err.Error()
	}

	updateResult := a.UpdateSubscription(subID)
	if strings.HasPrefix(updateResult, "Updated:") {
		return updateResult
	}

	// Roll back newly created subscription when initial sync failed.
	rolledBack := make([]Subscription, 0, len(a.Subscriptions))
	for _, s := range a.Subscriptions {
		if s.ID != subID {
			rolledBack = append(rolledBack, s)
		}
	}
	a.Subscriptions = rolledBack
	if err := a.SaveSubscriptions(); err != nil {
		return updateResult + " (rollback failed: " + err.Error() + ")"
	}
	return updateResult
}

func (a *App) DeleteSubscription(subID string) {
	newSubs := []Subscription{}
	for _, s := range a.Subscriptions {
		if s.ID != subID {
			newSubs = append(newSubs, s)
		}
	}
	a.Subscriptions = newSubs
	if err := a.SaveSubscriptions(); err != nil {
		a.log("Failed to save subscriptions after delete: " + err.Error())
	}

	newProfs := []Profile{}
	for _, p := range a.Profiles {
		if p.SubscriptionID != subID {
			newProfs = append(newProfs, p)
		}
	}
	a.Profiles = newProfs
	if err := a.SaveProfiles(); err != nil {
		a.log("Failed to save profiles after subscription delete: " + err.Error())
	}
	a.invalidateLastProfileIfMissing()
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
	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("Download failed: HTTP %d", resp.StatusCode)
	}

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
	seenLinks := make(map[string]struct{})
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "vless://") {
			if _, exists := seenLinks[line]; exists {
				continue
			}
			seenLinks[line] = struct{}{}
			newLinks = append(newLinks, line)
		}
	}

	if len(newLinks) == 0 {
		return "No valid links found"
	}

	targetSub.UpdatedAt = time.Now().Unix()

	existingByKey := make(map[string]Profile)
	for _, p := range a.Profiles {
		if p.SubscriptionID == subID {
			existingByKey[p.Key] = p
		}
	}

	tempProfiles := []Profile{}
	for _, p := range a.Profiles {
		if p.SubscriptionID != subID {
			tempProfiles = append(tempProfiles, p)
		}
	}
	a.Profiles = tempProfiles

	profilesCount := 0
	for _, link := range newLinks {
		u, err := url.Parse(link)
		if err != nil {
			continue
		}
		name := u.Fragment
		if name == "" {
			name = u.Hostname()
		}
		name, _ = url.QueryUnescape(name)

		id := uuid.New().String()
		createdAt := time.Now().Unix()
		if existing, ok := existingByKey[link]; ok {
			id = existing.ID
			createdAt = existing.CreatedAt
		}

		a.Profiles = append(a.Profiles, Profile{
			ID:             id,
			Name:           name,
			Key:            link,
			SubscriptionID: subID,
			CreatedAt:      createdAt,
		})
		profilesCount++
	}

	if profilesCount == 0 {
		return "No valid links found"
	}

	if err := a.SaveSubscriptions(); err != nil {
		return "Failed to save subscriptions: " + err.Error()
	}
	if err := a.SaveProfiles(); err != nil {
		return "Failed to save profiles: " + err.Error()
	}
	a.invalidateLastProfileIfMissing()

	return fmt.Sprintf("Updated: %d profiles", profilesCount)
}

func (a *App) GetSubscriptions() []Subscription {
	return a.LoadSubscriptions()
}
