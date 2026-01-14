package main

import (
	"bufio"
	"bytes"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

func (a *App) GetRunningProcesses() []string {
	processes := make(map[string]struct{})
	
	if runtime.GOOS == "windows" {
		cmd := exec.Command("tasklist", "/FO", "CSV", "/NH")
		a.configureCmd(cmd)
		
		output, err := cmd.Output()
		if err == nil {
			r := bytes.NewReader(output)
			scanner := bufio.NewScanner(r)
			for scanner.Scan() {
				line := scanner.Text()
				parts := strings.Split(line, "\",\"")
				if len(parts) > 0 {
					name := strings.Trim(parts[0], "\"")
					if !isSystemProcess(name) {
						processes[name] = struct{}{}
					}
				}
			}
		}
	} else {
		args := []string{"-e", "-o", "comm="}
		if runtime.GOOS == "darwin" {
			args = []string{"-A", "-o", "comm="}
		}

		cmd := exec.Command("ps", args...)
		a.configureCmd(cmd)
		
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				rawName := strings.TrimSpace(line)
				if rawName == "" {
					continue
				}

				name := filepath.Base(rawName)

				if !isSystemProcess(name) {
					processes[name] = struct{}{}
				}
			}
		}
	}

	result := make([]string, 0, len(processes))
	for name := range processes {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

func isSystemProcess(name string) bool {
	lower := strings.ToLower(name)

	if runtime.GOOS == "darwin" {
		if strings.Contains(lower, "helper") || 
		   strings.Contains(lower, "renderer") || 
		   strings.Contains(lower, "gpu") || 
		   strings.Contains(lower, "plugin") ||
		   strings.Contains(lower, "xpc") ||
		   strings.Contains(lower, "service") {
			return true
		}
	}

	windowsSys := []string{
		"system", "system idle process", "registry", "memcompression",
		"smss.exe", "csrss.exe", "wininit.exe", "services.exe", "lsass.exe",
		"svchost.exe", "fontdrvhost.exe", "winlogon.exe", "dwm.exe",
		"spoolsv.exe", "searchindexer.exe", "taskhostw.exe", "explorer.exe",
		"runtimebroker.exe", "shellexperiencehost.exe", "applicationframehost.exe",
		"dllhost.exe", "conhost.exe", "ctfmon.exe", "smartscreen.exe",
		"sihost.exe", "werfault.exe", "wudfhost.exe", "securityhealthservice.exe",
		"sgrmbroker.exe", "dasuan.exe",
	}

	for _, sys := range windowsSys {
		if lower == sys {
			return true
		}
	}

	if runtime.GOOS != "windows" {
		if strings.HasPrefix(name, "[") && strings.HasSuffix(name, "]") {
			return true
		}

		linuxSys := []string{
			"launchd", "kernel_task", "logd", "userEventAgent", "distnoted",
			"cfprefsd", "xpcproxy", "tccd", "com.apple", "mds", "mds_stores",
			"nsurlsessiond", "syslogd", "systemstats", "configd", "powerd",
			"lsd", "pkd", "secinitd", "trustd", "fseventsd", "diskarbitrationd",
			"systemd", "kthreadd", "rcu_sched", "rcu_bh", "migration",
			"watchdog", "ksoftirqd", "kworker", "dbus-daemon", "networkmanager",
			"polkitd", "wpa_supplicant", "avahi-daemon", "systemd-journal",
			"systemd-udevd", "systemd-logind", "systemd-resolv", "zsh", "bash", "sh",
			"login",
		}
		for _, sys := range linuxSys {
			if strings.EqualFold(lower, sys) || strings.HasPrefix(lower, sys) {
				return true
			}
		}
	}

	return false
}
