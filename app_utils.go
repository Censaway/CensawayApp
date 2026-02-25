package main

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type GithubRelease struct {
	TagName string               `json:"tag_name"`
	Assets  []GithubReleaseAsset `json:"assets"`
}

type GithubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Digest             string `json:"digest"`
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

	release, err := a.fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to get latest release info: %v", err)
	}
	version := release.TagName

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
	asset, ok := findReleaseAssetByName(release.Assets, fileName)
	if !ok {
		return fmt.Errorf("release asset not found: %s", fileName)
	}

	tempPath := filepath.Join(os.TempDir(), fileName)
	if err := a.downloadFile(asset.BrowserDownloadURL, tempPath); err != nil {
		return err
	}
	defer os.Remove(tempPath)

	expectedHash, verificationSource, err := a.expectedSHA256ForAsset(asset, release.Assets)
	if err != nil {
		return fmt.Errorf("failed to get expected checksum: %v", err)
	}

	actualHash, err := sha256File(tempPath)
	if err != nil {
		return fmt.Errorf("failed to hash downloaded file: %v", err)
	}

	if !strings.EqualFold(expectedHash, actualHash) {
		return fmt.Errorf("checksum mismatch for %s", fileName)
	}
	wailsRuntime.EventsEmit(a.ctx, "log", "Checksum verified ("+verificationSource+")")

	binDir := filepath.Join(a.getAppDataDir(), "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %v", err)
	}

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
	if err != nil {
		return fmt.Errorf("failed to locate extracted binary: %v", err)
	}

	dirEntries, err := os.ReadDir(binDir)
	if err != nil {
		return fmt.Errorf("failed to read bin directory: %v", err)
	}
	for _, entry := range dirEntries {
		if entry.IsDir() {
			if err := os.RemoveAll(filepath.Join(binDir, entry.Name())); err != nil {
				return fmt.Errorf("failed to clean extracted directory %s: %v", entry.Name(), err)
			}
		}
	}

	finalBin := filepath.Join(binDir, targetBinName)
	if _, err := os.Stat(finalBin); err != nil {
		return fmt.Errorf("final core binary not found: %v", err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(finalBin, 0755); err != nil {
			return fmt.Errorf("failed to set executable permissions: %v", err)
		}
	}

	wailsRuntime.EventsEmit(a.ctx, "log", "Core installed successfully.")
	return nil
}

func (a *App) fetchLatestRelease() (*GithubRelease, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/SagerNet/sing-box/releases/latest", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("github api returned status: %d", resp.StatusCode)
	}

	var release GithubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	if release.TagName == "" {
		return nil, fmt.Errorf("empty tag name in response")
	}

	return &release, nil
}

func (a *App) expectedSHA256ForAsset(asset *GithubReleaseAsset, assets []GithubReleaseAsset) (string, string, error) {
	if asset == nil {
		return "", "", errors.New("release asset is nil")
	}

	if hash, ok := parseSHA256Digest(asset.Digest); ok {
		return hash, "release asset digest", nil
	}

	checksumAsset, ok := findChecksumAsset(assets)
	if !ok {
		return "", "", errors.New("release has no sha256 digest and no checksums asset")
	}

	checksumText, err := a.downloadTextFile(checksumAsset.BrowserDownloadURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to download checksums: %w", err)
	}

	hash, err := checksumForFile(checksumText, asset.Name)
	if err != nil {
		return "", "", fmt.Errorf("checksum for %s not found: %w", asset.Name, err)
	}

	return hash, "checksums file " + checksumAsset.Name, nil
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

func (a *App) downloadTextFile(url string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "CensawayApp")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("download failed, status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func findReleaseAssetByName(assets []GithubReleaseAsset, name string) (*GithubReleaseAsset, bool) {
	for i := range assets {
		if assets[i].Name == name {
			return &assets[i], true
		}
	}

	lowerName := strings.ToLower(name)
	for i := range assets {
		if strings.ToLower(assets[i].Name) == lowerName {
			return &assets[i], true
		}
	}

	return nil, false
}

func findChecksumAsset(assets []GithubReleaseAsset) (*GithubReleaseAsset, bool) {
	for i := range assets {
		name := strings.ToLower(assets[i].Name)
		if strings.Contains(name, "sha256") &&
			(strings.Contains(name, "sum") || strings.Contains(name, "checksum")) &&
			(strings.HasSuffix(name, ".txt") || strings.HasSuffix(name, ".sha256") || strings.HasSuffix(name, ".sha256sum")) {
			return &assets[i], true
		}
	}

	for i := range assets {
		name := strings.ToLower(assets[i].Name)
		if strings.Contains(name, "sha256") && (strings.Contains(name, "sum") || strings.Contains(name, "checksum")) {
			return &assets[i], true
		}
	}

	return nil, false
}

func checksumForFile(content, filename string) (string, error) {
	targetName := strings.ToLower(filepath.Base(filename))

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if hash, name, ok := parseChecksumLine(line); ok {
			if strings.ToLower(filepath.Base(name)) == targetName {
				return hash, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", errors.New("checksum entry not found")
}

func parseChecksumLine(line string) (string, string, bool) {
	if idx := strings.Index(line, "="); idx != -1 {
		left := strings.TrimSpace(line[:idx])
		right := strings.TrimSpace(line[idx+1:])
		if isSHA256Hex(right) {
			openIdx := strings.Index(left, "(")
			closeIdx := strings.LastIndex(left, ")")
			if openIdx != -1 && closeIdx > openIdx {
				name := strings.TrimSpace(left[openIdx+1 : closeIdx])
				return strings.ToLower(right), name, true
			}
			return strings.ToLower(right), left, true
		}
	}

	parts := strings.Fields(line)
	if len(parts) >= 2 && isSHA256Hex(parts[0]) {
		name := strings.Join(parts[1:], " ")
		name = strings.TrimPrefix(name, "*")
		return strings.ToLower(parts[0]), strings.TrimSpace(name), true
	}

	return "", "", false
}

func parseSHA256Digest(digest string) (string, bool) {
	digest = strings.TrimSpace(digest)
	if digest == "" {
		return "", false
	}

	if idx := strings.Index(digest, ":"); idx != -1 {
		alg := strings.ToLower(strings.TrimSpace(digest[:idx]))
		value := strings.TrimSpace(digest[idx+1:])
		if alg != "sha256" || !isSHA256Hex(value) {
			return "", false
		}
		return strings.ToLower(value), true
	}

	if isSHA256Hex(digest) {
		return strings.ToLower(digest), true
	}

	return "", false
}

func isSHA256Hex(s string) bool {
	if len(s) != 64 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
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
		if _, err := io.Copy(outFile, rc); err != nil {
			outFile.Close()
			rc.Close()
			return err
		}
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
