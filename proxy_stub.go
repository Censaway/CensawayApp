//go:build !windows

package main

func (a *App) setSystemProxy(enable bool, port int) error {
	return nil
}

func (a *App) ensureProxyDisabled() {
}
