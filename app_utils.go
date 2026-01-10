package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type GithubRelease struct {
	TagName string `json:"tag_name"`
}

func (a *App) ensureWintun() error {
	if runtime.GOOS != "windows" {
		return nil
	}

	binDir := filepath.Join(a.getAppDataDir(), "bin")

	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin dir: %w", err)
	}

	dllPath := filepath.Join(binDir, "wintun.dll")
	if _, err := os.Stat(dllPath); err == nil {
		return nil
	}

	if len(wintunDll) == 0 {
		return fmt.Errorf("embedded wintun.dll is empty")
	}

	a.log("Extracting wintun.dll...")
	return os.WriteFile(dllPath, wintunDll, 0755)
}

func (a *App) getProxyBin() (string, error) {
	binName := "sing-box"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	localBin := filepath.Join(a.getAppDataDir(), "bin", binName)
	if _, err := os.Stat(localBin); err == nil {
		return localBin, nil
	}

	return "", fmt.Errorf("core_missing")
}

func (a *App) checkAndInstallCore() error {
	if _, err := a.getProxyBin(); err == nil {
		return nil
	}

	wailsRuntime.EventsEmit(a.ctx, "log", "Core missing. Fetching latest version info...")

	version, err := a.fetchLatestVersionTag()
	if err != nil {
		return fmt.Errorf("failed to get latest version: %v", err)
	}

	cleanVersion := strings.TrimPrefix(version, "v")
	wailsRuntime.EventsEmit(a.ctx, "log", fmt.Sprintf("Downloading Sing-box %s...", cleanVersion))

	osName := runtime.GOOS
	arch := runtime.GOARCH

	if arch == "amd64" {
		arch = "amd64"
	} else if arch == "arm64" {
		arch = "arm64"
	} else {
		return fmt.Errorf("unsupported architecture: %s", arch)
	}

	ext := "tar.gz"
	if osName == "windows" {
		ext = "zip"
	}

	if osName == "darwin" {
		osName = "darwin"
	}

	fileName := fmt.Sprintf("sing-box-%s-%s-%s.%s", cleanVersion, osName, arch, ext)
	downloadUrl := fmt.Sprintf("https://github.com/sagernet/sing-box/releases/download/%s/%s", version, fileName)

	tempPath := filepath.Join(os.TempDir(), fileName)
	if err := a.downloadFile(downloadUrl, tempPath); err != nil {
		return err
	}
	defer os.Remove(tempPath)

	binDir := filepath.Join(a.getAppDataDir(), "bin")
	os.MkdirAll(binDir, 0755)

	wailsRuntime.EventsEmit(a.ctx, "log", "Extracting...")
	if osName == "windows" {
		err = unzip(tempPath, binDir)
	} else {
		err = untar(tempPath, binDir)
	}
	if err != nil {
		return fmt.Errorf("extraction failed: %v", err)
	}

	targetBinName := "sing-box"
	if runtime.GOOS == "windows" {
		targetBinName += ".exe"
	}

	err = filepath.Walk(binDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == targetBinName {
			finalPath := filepath.Join(binDir, targetBinName)
			if path == finalPath {
				return nil
			}
			return os.Rename(path, finalPath)
		}
		return nil
	})

	dirEntries, _ := os.ReadDir(binDir)
	for _, entry := range dirEntries {
		if entry.IsDir() {
			os.RemoveAll(filepath.Join(binDir, entry.Name()))
		}
	}

	finalBin := filepath.Join(binDir, targetBinName)
	if runtime.GOOS != "windows" {
		os.Chmod(finalBin, 0755)
	}

	wailsRuntime.EventsEmit(a.ctx, "log", "Core installed successfully.")
	return nil
}

func (a *App) fetchLatestVersionTag() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/SagerNet/sing-box/releases/latest", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("github api returned status: %d", resp.StatusCode)
	}

	var release GithubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	if release.TagName == "" {
		return "", fmt.Errorf("empty tag name in response")
	}

	return release.TagName, nil
}

func (a *App) downloadFile(url string, dest string) error {
	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed, status: %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("zip slip attempt")
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
	}
	return nil
}

func untar(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		target := filepath.Join(dest, header.Name)
		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("tar slip attempt")
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
			os.Chmod(target, os.FileMode(header.Mode))
		}
	}
	return nil
}

func (a *App) ensurePermissions(binPath string) error {
	if runtime.GOOS == "windows" {
		return nil
	}

	if runtime.GOOS == "darwin" {
		exec.Command("xattr", "-d", "com.apple.quarantine", binPath).Run()

		info, err := os.Stat(binPath)
		if err == nil {
			if (info.Mode() & os.ModeSetuid) != 0 {
				return nil
			}
		}

		if a.ctx != nil {
			wailsRuntime.EventsEmit(a.ctx, "log", "Requesting admin rights (osascript)...")
		}

		cmdStr := fmt.Sprintf("chown root:admin \\\"%s\\\" && chmod +s \\\"%s\\\"", binPath, binPath)
		script := fmt.Sprintf("do shell script \"%s\" with administrator privileges", cmdStr)

		err = exec.Command("osascript", "-e", script).Run()
		if err != nil {
			return fmt.Errorf("failed to set SUID: %v", err)
		}
		return nil
	}

	checkCmd := exec.Command("getcap", binPath)
	out, _ := checkCmd.CombinedOutput()
	if strings.Contains(string(out), "cap_net_admin") {
		return nil
	}

	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, "log", "Requesting admin rights (pkexec)...")
	}

	err := exec.Command("pkexec", "setcap", "cap_net_admin,cap_net_bind_service=+ep", binPath).Run()
	if err != nil {
		return exec.Command("sudo", "setcap", "cap_net_admin,cap_net_bind_service=+ep", binPath).Run()
	}
	return nil
}
