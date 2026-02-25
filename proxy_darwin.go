//go:build darwin

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type darwinProxyConfig struct {
	Enabled bool   `json:"enabled"`
	Server  string `json:"server,omitempty"`
	Port    int    `json:"port,omitempty"`
}

type darwinProxyServiceSnapshot struct {
	Service string            `json:"service"`
	Web     darwinProxyConfig `json:"web"`
	Secure  darwinProxyConfig `json:"secure"`
	Socks   darwinProxyConfig `json:"socks"`
	Bypass  []string          `json:"bypass,omitempty"`
}

type darwinProxySnapshot struct {
	Services []darwinProxyServiceSnapshot `json:"services"`
}

func (a *App) darwinProxySnapshotPath() string {
	return filepath.Join(a.getAppDataDir(), "darwin_proxy_snapshot.json")
}

func (a *App) loadDarwinProxySnapshot() (*darwinProxySnapshot, error) {
	data, err := os.ReadFile(a.darwinProxySnapshotPath())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var snapshot darwinProxySnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (a *App) saveDarwinProxySnapshot(snapshot darwinProxySnapshot) error {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	return os.WriteFile(a.darwinProxySnapshotPath(), data, 0600)
}

func (a *App) clearDarwinProxySnapshot() error {
	err := os.Remove(a.darwinProxySnapshotPath())
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func runNetworksetup(args ...string) (string, error) {
	cmd := exec.Command("networksetup", args...)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		if output != "" {
			return "", fmt.Errorf("networksetup %s failed: %s", strings.Join(args, " "), output)
		}
		return "", fmt.Errorf("networksetup %s failed: %w", strings.Join(args, " "), err)
	}
	return output, nil
}

func parseNetworksetupKV(output string) map[string]string {
	values := map[string]string{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		idx := strings.Index(line, ":")
		if idx == -1 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(line[:idx]))
		value := strings.TrimSpace(line[idx+1:])
		values[key] = value
	}
	return values
}

func parseDarwinProxyConfig(output string) (darwinProxyConfig, error) {
	values := parseNetworksetupKV(output)

	cfg := darwinProxyConfig{}
	enabledRaw := strings.ToLower(strings.TrimSpace(values["enabled"]))
	cfg.Enabled = enabledRaw == "yes" || enabledRaw == "1" || enabledRaw == "on"
	cfg.Server = strings.TrimSpace(values["server"])

	portRaw := strings.TrimSpace(values["port"])
	if portRaw != "" {
		port, err := strconv.Atoi(portRaw)
		if err != nil {
			return cfg, fmt.Errorf("invalid proxy port %q", portRaw)
		}
		cfg.Port = port
	}

	return cfg, nil
}

func darwinListNetworkServices() ([]string, error) {
	output, err := runNetworksetup("-listallnetworkservices")
	if err != nil {
		return nil, err
	}

	services := []string{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(strings.ToLower(line), "an asterisk") {
			continue
		}
		if strings.HasPrefix(line, "*") {
			// Disabled service, skip.
			continue
		}
		services = append(services, line)
	}
	return services, nil
}

func darwinGetBypassDomains(service string) ([]string, error) {
	output, err := runNetworksetup("-getproxybypassdomains", service)
	if err != nil {
		return nil, err
	}
	if output == "" {
		return []string{}, nil
	}

	lines := strings.Split(output, "\n")
	domains := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if strings.Contains(lower, "there aren't any bypass domains") {
			return []string{}, nil
		}
		domains = append(domains, line)
	}
	return domains, nil
}

func darwinSetBypassDomains(service string, domains []string) error {
	args := []string{"-setproxybypassdomains", service}
	if len(domains) == 0 {
		args = append(args, "Empty")
	} else {
		args = append(args, domains...)
	}
	_, err := runNetworksetup(args...)
	return err
}

func readDarwinServiceSnapshot(service string) (darwinProxyServiceSnapshot, error) {
	snapshot := darwinProxyServiceSnapshot{Service: service}

	output, err := runNetworksetup("-getwebproxy", service)
	if err != nil {
		return snapshot, err
	}
	snapshot.Web, err = parseDarwinProxyConfig(output)
	if err != nil {
		return snapshot, err
	}

	output, err = runNetworksetup("-getsecurewebproxy", service)
	if err != nil {
		return snapshot, err
	}
	snapshot.Secure, err = parseDarwinProxyConfig(output)
	if err != nil {
		return snapshot, err
	}

	output, err = runNetworksetup("-getsocksfirewallproxy", service)
	if err != nil {
		return snapshot, err
	}
	snapshot.Socks, err = parseDarwinProxyConfig(output)
	if err != nil {
		return snapshot, err
	}

	bypass, err := darwinGetBypassDomains(service)
	if err != nil {
		return snapshot, err
	}
	snapshot.Bypass = bypass

	return snapshot, nil
}

func enableDarwinServiceProxy(service string, port int) error {
	portStr := strconv.Itoa(port)

	if _, err := runNetworksetup("-setwebproxy", service, "127.0.0.1", portStr); err != nil {
		return err
	}
	if _, err := runNetworksetup("-setwebproxystate", service, "on"); err != nil {
		return err
	}

	if _, err := runNetworksetup("-setsecurewebproxy", service, "127.0.0.1", portStr); err != nil {
		return err
	}
	if _, err := runNetworksetup("-setsecurewebproxystate", service, "on"); err != nil {
		return err
	}

	if _, err := runNetworksetup("-setsocksfirewallproxy", service, "127.0.0.1", portStr); err != nil {
		return err
	}
	if _, err := runNetworksetup("-setsocksfirewallproxystate", service, "on"); err != nil {
		return err
	}

	currentBypass, err := darwinGetBypassDomains(service)
	if err != nil {
		return err
	}
	targetBypass := append([]string{}, currentBypass...)
	targetBypass = append(targetBypass, "localhost", "127.0.0.1", "::1")

	uniq := make([]string, 0, len(targetBypass))
	seen := map[string]struct{}{}
	for _, entry := range targetBypass {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		key := strings.ToLower(entry)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		uniq = append(uniq, entry)
	}
	return darwinSetBypassDomains(service, uniq)
}

func isUnknownDarwinServiceError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "is not a recognized network service")
}

func restoreDarwinServiceProxy(snapshot darwinProxyServiceSnapshot) error {
	service := snapshot.Service

	if snapshot.Web.Enabled {
		if snapshot.Web.Server == "" || snapshot.Web.Port <= 0 {
			return errors.New("invalid snapshot for web proxy")
		}
		if _, err := runNetworksetup("-setwebproxy", service, snapshot.Web.Server, strconv.Itoa(snapshot.Web.Port)); err != nil {
			return err
		}
		if _, err := runNetworksetup("-setwebproxystate", service, "on"); err != nil {
			return err
		}
	} else {
		if _, err := runNetworksetup("-setwebproxystate", service, "off"); err != nil {
			return err
		}
	}

	if snapshot.Secure.Enabled {
		if snapshot.Secure.Server == "" || snapshot.Secure.Port <= 0 {
			return errors.New("invalid snapshot for secure web proxy")
		}
		if _, err := runNetworksetup("-setsecurewebproxy", service, snapshot.Secure.Server, strconv.Itoa(snapshot.Secure.Port)); err != nil {
			return err
		}
		if _, err := runNetworksetup("-setsecurewebproxystate", service, "on"); err != nil {
			return err
		}
	} else {
		if _, err := runNetworksetup("-setsecurewebproxystate", service, "off"); err != nil {
			return err
		}
	}

	if snapshot.Socks.Enabled {
		if snapshot.Socks.Server == "" || snapshot.Socks.Port <= 0 {
			return errors.New("invalid snapshot for socks proxy")
		}
		if _, err := runNetworksetup("-setsocksfirewallproxy", service, snapshot.Socks.Server, strconv.Itoa(snapshot.Socks.Port)); err != nil {
			return err
		}
		if _, err := runNetworksetup("-setsocksfirewallproxystate", service, "on"); err != nil {
			return err
		}
	} else {
		if _, err := runNetworksetup("-setsocksfirewallproxystate", service, "off"); err != nil {
			return err
		}
	}

	return darwinSetBypassDomains(service, snapshot.Bypass)
}

func (a *App) setSystemProxy(enable bool, port int) error {
	if enable && (port < 1 || port > 65535) {
		return fmt.Errorf("invalid system proxy port: %d", port)
	}

	services, err := darwinListNetworkServices()
	if err != nil {
		return err
	}

	if enable {
		if len(services) == 0 {
			return errors.New("no active network services found")
		}

		snapshot, err := a.loadDarwinProxySnapshot()
		if err != nil {
			return err
		}

		if snapshot == nil {
			captured := darwinProxySnapshot{
				Services: make([]darwinProxyServiceSnapshot, 0, len(services)),
			}
			for _, service := range services {
				state, err := readDarwinServiceSnapshot(service)
				if err != nil {
					return err
				}
				captured.Services = append(captured.Services, state)
			}
			if err := a.saveDarwinProxySnapshot(captured); err != nil {
				return err
			}
		}

		for _, service := range services {
			if err := enableDarwinServiceProxy(service, port); err != nil {
				return err
			}
		}
		return nil
	}

	snapshot, err := a.loadDarwinProxySnapshot()
	if err != nil {
		return err
	}
	if snapshot == nil {
		return nil
	}

	for _, service := range snapshot.Services {
		if err := restoreDarwinServiceProxy(service); err != nil {
			if isUnknownDarwinServiceError(err) {
				continue
			}
			return err
		}
	}

	return a.clearDarwinProxySnapshot()
}

func (a *App) ensureProxyDisabled() {
	_ = a.setSystemProxy(false, 0)
}
