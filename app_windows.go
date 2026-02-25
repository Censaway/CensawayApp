//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func (a *App) configureCmd(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
}

func (a *App) EnableAutostart() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	exe, err := os.Executable()
	if err != nil {
		return err
	}

	exe, err = filepath.Abs(exe)
	if err != nil {
		return err
	}

	return k.SetStringValue("CensawayApp", "\""+exe+"\" -tray")
}

func (a *App) DisableAutostart() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	err = k.DeleteValue("CensawayApp")
	if err == nil ||
		errors.Is(err, registry.ErrNotExist) ||
		errors.Is(err, syscall.ERROR_FILE_NOT_FOUND) ||
		errors.Is(err, syscall.ERROR_PATH_NOT_FOUND) {
		return nil
	}
	return err
}

func (a *App) platformInit() error {
	return nil
}

func (a *App) cleanupZombies() {
	info, err := a.loadCoreProcessInfo()
	if err != nil {
		a.log("Failed to read core process info: " + err.Error())
		return
	}
	if info == nil {
		return
	}
	if info.PID <= 0 {
		a.clearCoreProcessInfo()
		return
	}
	if !windowsPIDLooksLikeSingBox(info.PID) {
		a.clearCoreProcessInfo()
		return
	}

	cmd := exec.Command("taskkill", "/F", "/PID", strconv.Itoa(info.PID))
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
	_ = cmd.Run()
	a.clearCoreProcessInfo()
}

func windowsPIDLooksLikeSingBox(pid int) bool {
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/FO", "CSV", "/NH")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
	out, err := cmd.Output()
	if err != nil {
		return false
	}

	line := strings.ToLower(strings.TrimSpace(string(out)))
	if line == "" || strings.Contains(line, "no tasks are running") {
		return false
	}

	return strings.Contains(line, "sing-box.exe")
}

func (a *App) ensurePermissions(binPath string) error {
	if a.Settings.RunMode == "proxy" {
		return nil
	}

	if !a.checkAdmin() {
		a.log("Admin rights missing for TUN mode. Requesting elevation...")

		err := a.runMeElevated()
		if err != nil {
			return fmt.Errorf("failed to elevate: %v", err)
		}

		os.Exit(0)
		return nil
	}
	return nil
}

func (a *App) checkAdmin() bool {
	var sid *windows.SID
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)

	token := windows.Token(0)
	member, err := token.IsMember(sid)
	if err != nil {
		return false
	}
	return member
}

func (a *App) runMeElevated() error {
	verb := "runas"
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := strings.Join(os.Args[1:], " ")

	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argsPtr, _ := syscall.UTF16PtrFromString(args)

	var showCmd int32 = 1

	err := windows.ShellExecute(0, verbPtr, exePtr, argsPtr, cwdPtr, showCmd)
	if err != nil {
		return err
	}
	return nil
}
