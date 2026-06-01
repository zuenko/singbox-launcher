package core

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"singbox-launcher/internal/debuglog"
	"singbox-launcher/internal/platform"
)

// DownloadXrayCore downloads and installs Xray-core from GitHub releases.
// Version "" means "latest".
func (ac *AppController) DownloadXrayCore(ctx context.Context, version string, progressChan chan DownloadProgress) {
	defer close(progressChan)

	progressChan <- DownloadProgress{Progress: 5, Message: "Getting Xray release info...", Status: "downloading"}

	releaseURL := "https://api.github.com/repos/XTLS/Xray-core/releases/latest"
	if version != "" {
		releaseURL = fmt.Sprintf("https://api.github.com/repos/XTLS/Xray-core/releases/tags/%s", version)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", releaseURL, nil)
	if err != nil {
		progressChan <- DownloadProgress{Progress: 0, Message: fmt.Sprintf("Failed to create request: %v", err), Status: "error", Error: err}
		return
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "singbox-launcher")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		progressChan <- DownloadProgress{Progress: 0, Message: fmt.Sprintf("Failed to get release info: %v", err), Status: "error", Error: err}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		progressChan <- DownloadProgress{Progress: 0, Message: fmt.Sprintf("GitHub API returned %d", resp.StatusCode), Status: "error", Error: fmt.Errorf("HTTP %d", resp.StatusCode)}
		return
	}

	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
			Size               int64  `json:"size"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		progressChan <- DownloadProgress{Progress: 0, Message: fmt.Sprintf("Failed to parse release info: %v", err), Status: "error", Error: err}
		return
	}

	// Find correct asset
	assetName := xrayAssetName()
	var assetURL string
	for _, a := range release.Assets {
		if strings.EqualFold(a.Name, assetName) {
			assetURL = a.BrowserDownloadURL
			break
		}
	}
	if assetURL == "" {
		progressChan <- DownloadProgress{Progress: 0, Message: fmt.Sprintf("Asset %s not found in release %s", assetName, release.TagName), Status: "error", Error: fmt.Errorf("asset not found")}
		return
	}

	progressChan <- DownloadProgress{Progress: 15, Message: fmt.Sprintf("Downloading %s...", assetName), Status: "downloading"}

	tempDir := filepath.Join(ac.FileService.ExecDir, "temp")
	_ = os.MkdirAll(tempDir, platform.DefaultDirMode)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	archivePath := filepath.Join(tempDir, assetName)
	if err := ac.downloadFile(ctx, assetURL, archivePath, progressChan); err != nil {
		progressChan <- DownloadProgress{Progress: 0, Message: fmt.Sprintf("Download failed: %v", err), Status: "error", Error: err}
		return
	}

	progressChan <- DownloadProgress{Progress: 80, Message: "Extracting archive...", Status: "extracting"}

	binDir := filepath.Join(ac.FileService.ExecDir, "bin")
	_ = os.MkdirAll(binDir, platform.DefaultDirMode)

	if err := extractXrayArchive(archivePath, tempDir, binDir); err != nil {
		progressChan <- DownloadProgress{Progress: 0, Message: fmt.Sprintf("Extraction failed: %v", err), Status: "error", Error: err}
		return
	}

	progressChan <- DownloadProgress{Progress: 100, Message: fmt.Sprintf("Xray %s installed successfully", release.TagName), Status: "done"}
}

// xrayAssetName returns the expected Xray release asset name for current platform.
func xrayAssetName() string {
	var osName, arch string
	switch runtime.GOOS {
	case "windows":
		osName = "windows"
	case "linux":
		osName = "linux"
	case "darwin":
		osName = "macos"
	default:
		osName = runtime.GOOS
	}

	switch runtime.GOARCH {
	case "amd64":
		arch = "64"
	case "arm64":
		arch = "arm64-v8a"
	case "386":
		arch = "32"
	default:
		arch = runtime.GOARCH
	}

	return fmt.Sprintf("Xray-%s-%s.zip", osName, arch)
}

// extractXrayArchive extracts xray.exe/xray from the zip to binDir.
func extractXrayArchive(archivePath, tempDir, binDir string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()

	var targetName string
	if runtime.GOOS == "windows" {
		targetName = "xray.exe"
	} else {
		targetName = "xray"
	}

	for _, f := range r.File {
		if strings.EqualFold(filepath.Base(f.Name), targetName) {
			rc, err := f.Open()
			if err != nil {
				return fmt.Errorf("open file in zip: %w", err)
			}
			defer rc.Close()

			outPath := filepath.Join(binDir, targetName)
			out, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return fmt.Errorf("create output file: %w", err)
			}
			defer out.Close()

			_, err = io.Copy(out, rc)
			if err != nil {
				return fmt.Errorf("copy binary: %w", err)
			}
			debuglog.InfoLog("Xray downloader: extracted %s to %s", targetName, outPath)
			return nil
		}
	}

	return fmt.Errorf("%s not found in archive", targetName)
}
