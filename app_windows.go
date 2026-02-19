//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

	return k.DeleteValue("CensawayApp")
}

func (a *App) platformInit() error {
	return nil
}

func (a *App) cleanupZombies() {
	cmd := exec.Command("taskkill", "/F", "/IM", "sing-box.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}
	_ = cmd.Run()
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