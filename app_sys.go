package main

import (
	"bufio"
	"bytes"
	"os/exec"
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
		cmd := exec.Command("ps", "-e", "-o", "comm=")
		a.configureCmd(cmd)
		
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				name := strings.TrimSpace(line)
				if name != "" && !isSystemProcess(name) {
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
			"systemd", "kthreadd", "rcu_sched", "rcu_bh", "migration",
			"watchdog", "ksoftirqd", "kworker", "dbus-daemon", "networkmanager",
			"polkitd", "wpa_supplicant", "avahi-daemon", "systemd-journal",
			"systemd-udevd", "systemd-logind", "systemd-resolv",
		}
		for _, sys := range linuxSys {
			if strings.HasPrefix(lower, sys) {
				return true
			}
		}
	}

	return false
}
