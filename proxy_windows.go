//go:build windows

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

var (
	modwininet            = syscall.NewLazyDLL("wininet.dll")
	procInternetSetOption = modwininet.NewProc("InternetSetOptionW")
)

const (
	INTERNET_OPTION_SETTINGS_CHANGED = 39
	INTERNET_OPTION_REFRESH          = 37
)

type windowsProxySnapshot struct {
	ProxyEnable      uint32 `json:"proxy_enable"`
	ProxyServer      string `json:"proxy_server,omitempty"`
	ProxyServerSet   bool   `json:"proxy_server_set"`
	ProxyOverride    string `json:"proxy_override,omitempty"`
	ProxyOverrideSet bool   `json:"proxy_override_set"`
}

func (a *App) proxySnapshotPath() string {
	return filepath.Join(a.getAppDataDir(), "windows_proxy_snapshot.json")
}

func (a *App) loadProxySnapshot() (*windowsProxySnapshot, error) {
	data, err := os.ReadFile(a.proxySnapshotPath())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var snapshot windowsProxySnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func (a *App) saveProxySnapshot(snapshot windowsProxySnapshot) error {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	return os.WriteFile(a.proxySnapshotPath(), data, 0600)
}

func (a *App) clearProxySnapshot() error {
	err := os.Remove(a.proxySnapshotPath())
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func isMissingRegistryValue(err error) bool {
	return errors.Is(err, registry.ErrNotExist) ||
		errors.Is(err, syscall.ERROR_FILE_NOT_FOUND) ||
		errors.Is(err, syscall.ERROR_PATH_NOT_FOUND)
}

func readProxySnapshotFromRegistry(k registry.Key) (windowsProxySnapshot, error) {
	snapshot := windowsProxySnapshot{}

	if val, _, err := k.GetIntegerValue("ProxyEnable"); err == nil {
		snapshot.ProxyEnable = uint32(val)
	} else if !isMissingRegistryValue(err) {
		return snapshot, err
	}

	if val, _, err := k.GetStringValue("ProxyServer"); err == nil {
		snapshot.ProxyServer = val
		snapshot.ProxyServerSet = true
	} else if !isMissingRegistryValue(err) {
		return snapshot, err
	}

	if val, _, err := k.GetStringValue("ProxyOverride"); err == nil {
		snapshot.ProxyOverride = val
		snapshot.ProxyOverrideSet = true
	} else if !isMissingRegistryValue(err) {
		return snapshot, err
	}

	return snapshot, nil
}

func restoreProxySnapshotToRegistry(k registry.Key, snapshot windowsProxySnapshot) error {
	if err := k.SetDWordValue("ProxyEnable", snapshot.ProxyEnable); err != nil {
		return err
	}

	if snapshot.ProxyServerSet {
		if err := k.SetStringValue("ProxyServer", snapshot.ProxyServer); err != nil {
			return err
		}
	} else {
		_ = k.DeleteValue("ProxyServer")
	}

	if snapshot.ProxyOverrideSet {
		if err := k.SetStringValue("ProxyOverride", snapshot.ProxyOverride); err != nil {
			return err
		}
	} else {
		_ = k.DeleteValue("ProxyOverride")
	}

	return nil
}

func (a *App) setSystemProxy(enable bool, port int) error {
	k, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
		registry.QUERY_VALUE|registry.SET_VALUE,
	)
	if err != nil {
		return err
	}
	defer k.Close()

	if enable {
		snapshot, err := a.loadProxySnapshot()
		if err != nil {
			return err
		}
		if snapshot == nil {
			currentSnapshot, err := readProxySnapshotFromRegistry(k)
			if err != nil {
				return err
			}
			if err := a.saveProxySnapshot(currentSnapshot); err != nil {
				return err
			}
		}

		if err := k.SetDWordValue("ProxyEnable", 1); err != nil {
			return err
		}
		if err := k.SetStringValue("ProxyServer", fmt.Sprintf("127.0.0.1:%d", port)); err != nil {
			return err
		}
		if err := k.SetStringValue("ProxyOverride", "<local>;localhost;127.*;10.*;172.16.*;192.168.*"); err != nil {
			return err
		}
	} else {
		snapshot, err := a.loadProxySnapshot()
		if err != nil {
			return err
		}
		if snapshot == nil {
			return nil
		}

		if err := restoreProxySnapshotToRegistry(k, *snapshot); err != nil {
			return err
		}
		if err := a.clearProxySnapshot(); err != nil {
			return err
		}
	}

	procInternetSetOption.Call(0, INTERNET_OPTION_SETTINGS_CHANGED, 0, 0)
	procInternetSetOption.Call(0, INTERNET_OPTION_REFRESH, 0, 0)

	return nil
}

func (a *App) ensureProxyDisabled() {
	_ = a.setSystemProxy(false, 0)
}
