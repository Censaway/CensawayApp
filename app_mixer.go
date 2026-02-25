package main

import (
	"encoding/json"
	"os"

	"github.com/google/uuid"
)

func (a *App) LoadMixedProfiles() []MixedProfile {
	data, err := os.ReadFile(a.getMixedProfilesPath())
	if err != nil {
		return []MixedProfile{}
	}
	json.Unmarshal(data, &a.MixedProfiles)
	if a.MixedProfiles == nil {
		a.MixedProfiles = []MixedProfile{}
	}
	return a.MixedProfiles
}

func (a *App) SaveMixedProfiles() error {
	data, err := json.MarshalIndent(a.MixedProfiles, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(a.getMixedProfilesPath(), data, 0600)
}

func (a *App) CreateMixedProfile(name string, relayID string, exitID string) string {
	if name == "" {
		name = "New Chain"
	}

	newMix := MixedProfile{
		ID:      uuid.New().String(),
		Name:    name,
		RelayID: relayID,
		ExitID:  exitID,
	}

	a.MixedProfiles = append(a.MixedProfiles, newMix)
	if err := a.SaveMixedProfiles(); err != nil {
		return "Save failed: " + err.Error()
	}
	return "OK"
}

func (a *App) DeleteMixedProfile(id string) {
	newMix := []MixedProfile{}
	for _, m := range a.MixedProfiles {
		if m.ID != id {
			newMix = append(newMix, m)
		}
	}
	a.MixedProfiles = newMix
	if err := a.SaveMixedProfiles(); err != nil {
		a.log("Failed to save mixed profiles after delete: " + err.Error())
	}
}

func (a *App) UpdateMixedProfile(id string, name string, relayID string, exitID string) string {
	for i, m := range a.MixedProfiles {
		if m.ID == id {
			a.MixedProfiles[i].Name = name
			a.MixedProfiles[i].RelayID = relayID
			a.MixedProfiles[i].ExitID = exitID
			if err := a.SaveMixedProfiles(); err != nil {
				return "Save failed: " + err.Error()
			}
			return "OK"
		}
	}
	return "Not Found"
}

func (a *App) GetMixedProfiles() []MixedProfile {
	return a.LoadMixedProfiles()
}
