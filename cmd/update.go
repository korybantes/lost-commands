package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	releaseRepoOwner = "korybantes"
	releaseRepoName  = "lost-commands"
)

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update lost to the latest GitHub release",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSelfUpdate()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func runSelfUpdate() error {
	assetName, err := currentAssetName()
	if err != nil {
		return err
	}

	release, err := fetchLatestRelease()
	if err != nil {
		return err
	}

	latestTag := strings.TrimSpace(release.TagName)
	if latestTag == "" {
		return fmt.Errorf("latest release does not include a tag name")
	}

	currentTag := appVersion
	if !strings.HasPrefix(currentTag, "v") {
		currentTag = "v" + currentTag
	}

	if latestTag == currentTag || strings.TrimPrefix(latestTag, "v") == strings.TrimPrefix(currentTag, "v") {
		fmt.Printf("lost is already up to date (%s)\n", latestTag)
		return nil
	}

	var assetURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			assetURL = asset.BrowserDownloadURL
			break
		}
	}
	if assetURL == "" {
		return fmt.Errorf("latest release %s does not include asset %q", latestTag, assetName)
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to locate current executable: %w", err)
	}

	tmpPath := filepath.Join(os.TempDir(), fmt.Sprintf("lost-update-%d%s", time.Now().UnixNano(), filepath.Ext(assetName)))
	if err := downloadAsset(assetURL, tmpPath); err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		_ = os.Chmod(tmpPath, 0755)
	}

	if runtime.GOOS == "windows" {
		if err := scheduleWindowsReplacement(exePath, tmpPath); err != nil {
			return err
		}
		fmt.Printf("Update to %s has been scheduled.\n", latestTag)
		fmt.Println("Please close this terminal and reopen it, then run: lost --version")
		return nil
	}

	if err := replaceExecutable(exePath, tmpPath); err != nil {
		if errors.Is(err, os.ErrPermission) || os.IsPermission(err) {
			return fmt.Errorf("permission denied updating %s (try running with sudo): %w", exePath, err)
		}
		return err
	}

	fmt.Printf("Updated lost from %s to %s\n", currentTag, latestTag)
	return nil
}

func currentAssetName() (string, error) {
	switch runtime.GOOS {
	case "windows":
		if runtime.GOARCH == "amd64" {
			return "lost-windows-amd64.exe", nil
		}
	case "darwin":
		if runtime.GOARCH == "amd64" {
			return "lost-darwin-amd64", nil
		}
		if runtime.GOARCH == "arm64" {
			return "lost-darwin-arm64", nil
		}
	case "linux":
		if runtime.GOARCH == "amd64" {
			return "lost-linux-amd64", nil
		}
		if runtime.GOARCH == "arm64" {
			return "lost-linux-arm64", nil
		}
	}
	return "", fmt.Errorf("unsupported platform for auto-update: %s/%s", runtime.GOOS, runtime.GOARCH)
}

func fetchLatestRelease() (*githubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", releaseRepoOwner, releaseRepoName)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "lost-updater")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query GitHub releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("no GitHub release found yet for %s/%s (create a release first)", releaseRepoOwner, releaseRepoName)
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("github API returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode GitHub release response: %w", err)
	}
	return &release, nil
}

func downloadAsset(url, destination string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "lost-updater")

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download update asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("asset download failed with status %s", resp.Status)
	}

	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return err
	}
	return nil
}

func replaceExecutable(exePath, downloadedPath string) error {
	exeDir := filepath.Dir(exePath)
	newPath := filepath.Join(exeDir, ".lost.new")
	backupPath := exePath + ".bak"

	_ = os.Remove(newPath)
	_ = os.Remove(backupPath)

	if err := copyFile(downloadedPath, newPath); err != nil {
		return err
	}
	if err := os.Chmod(newPath, 0755); err != nil {
		return err
	}

	if err := os.Rename(exePath, backupPath); err != nil {
		return err
	}
	if err := os.Rename(newPath, exePath); err != nil {
		_ = os.Rename(backupPath, exePath)
		return err
	}

	_ = os.Remove(backupPath)
	_ = os.Remove(downloadedPath)
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func scheduleWindowsReplacement(exePath, downloadedPath string) error {
	scriptPath := filepath.Join(os.TempDir(), fmt.Sprintf("lost-updater-%d.ps1", time.Now().UnixNano()))

	escape := func(s string) string {
		return strings.ReplaceAll(s, "'", "''")
	}

	script := fmt.Sprintf(`$ErrorActionPreference = "Stop"
$pidToWait = %d
$src = '%s'
$dst = '%s'

for ($i=0; $i -lt 120; $i++) {
  if (-not (Get-Process -Id $pidToWait -ErrorAction SilentlyContinue)) { break }
  Start-Sleep -Milliseconds 250
}

Copy-Item -LiteralPath $src -Destination $dst -Force
Remove-Item -LiteralPath $src -Force -ErrorAction SilentlyContinue
Remove-Item -LiteralPath $MyInvocation.MyCommand.Path -Force -ErrorAction SilentlyContinue
`, os.Getpid(), escape(downloadedPath), escape(exePath))

	if err := os.WriteFile(scriptPath, []byte(script), 0600); err != nil {
		return err
	}

	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}
