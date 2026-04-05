package xray

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	latestReleaseAPI = "https://api.github.com/repos/XTLS/Xray-core/releases/latest"
	downloadTpl      = "https://github.com/XTLS/Xray-core/releases/download/%s/%s"
)

type Manager struct {
	BinDir  string
	BinPath string
	Client  *http.Client
}

type releaseInfo struct {
	TagName string `json:"tag_name"`
}

func XrayManager(binDir string) *Manager {
	return &Manager{
		BinDir:  binDir,
		BinPath: filepath.Join(binDir, binaryName()),
		Client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Setup:
// - без localPath: скачать latest release и поставить/обновить
// - с localPath: поставить/обновить из локального zip или бинарника
func (m *Manager) Setup(localPath string) (string, error) {
	if err := os.MkdirAll(m.BinDir, 0o755); err != nil {
		return "", fmt.Errorf("create bin dir: %w", err)
	}

	alreadyInstalled := fileExists(m.BinPath)

	if localPath != "" {
		if err := m.setupFromLocal(localPath); err != nil {
			return "", err
		}
		if alreadyInstalled {
			return fmt.Sprintf("xray updated from local file: %s", localPath), nil
		}
		return fmt.Sprintf("xray installed from local file: %s", localPath), nil
	}

	if err := m.setupFromLatest(); err != nil {
		return "", err
	}

	if alreadyInstalled {
		ver := exec.Command("./bin/xray", "version")
		out, err := ver.Output()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Output: %s\n", out)
		return "xray updated from latest release", nil
	}
	ver := exec.Command("./bin/xray", "version")
	out, err := ver.Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Output: %s\n", out)
	return "xray installed from latest release", nil
}

func (m *Manager) Remove() (string, error) {
	if !fileExists(m.BinPath) {
		return "xray is not installed", nil
	}

	if err := os.Remove(m.BinPath); err != nil {
		return "", fmt.Errorf("remove xray: %w", err)
	}

	return fmt.Sprintf("xray removed: %s", m.BinPath), nil
}

func (m *Manager) setupFromLatest() error {
	version, err := m.latestVersion()
	if err != nil {
		return err
	}

	archive, err := archiveName()
	if err != nil {
		return err
	}

	url := fmt.Sprintf(downloadTpl, version, archive)
	fmt.Println(url)

	resp, err := m.Client.Get(url)
	if err != nil {
		return fmt.Errorf("download xray: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read archive: %w", err)
	}

	fmt.Printf("Unpacking to %s\n", m.BinPath)
	return unzipBinary(data, m.BinPath)
}

func (m *Manager) setupFromLocal(localPath string) error {
	info, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("local file error: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("expected file, got directory: %s", localPath)
	}

	if strings.EqualFold(filepath.Ext(localPath), ".zip") {
		data, err := os.ReadFile(localPath)
		if err != nil {
			return fmt.Errorf("read local zip: %w", err)
		}
		return unzipBinary(data, m.BinPath)
	}

	return copyFile(localPath, m.BinPath, 0o755)
}

func (m *Manager) latestVersion() (string, error) {
	req, err := http.NewRequest(http.MethodGet, latestReleaseAPI, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "proxy-cli")

	resp, err := m.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request github api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github api failed: %s", resp.Status)
	}

	var rel releaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", fmt.Errorf("decode release info: %w", err)
	}
	if rel.TagName == "" {
		return "", fmt.Errorf("empty release tag")
	}

	return rel.TagName, nil
}

func binaryName() string {
	if runtime.GOOS == "windows" {
		return "xray.exe"
	}
	return "xray"
}

func archiveName() (string, error) {
	switch runtime.GOOS {
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			return "Xray-linux-64.zip", nil
		case "386":
			return "Xray-linux-32.zip", nil
		case "arm64":
			return "Xray-linux-arm64-v8a.zip", nil
		case "arm":
			return "Xray-linux-arm32-v7a.zip", nil
		}
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			return "Xray-macos-64.zip", nil
		case "arm64":
			return "Xray-macos-arm64-v8a.zip", nil
		}
	case "windows":
		switch runtime.GOARCH {
		case "amd64":
			return "Xray-windows-64.zip", nil
		case "386":
			return "Xray-windows-32.zip", nil
		case "arm64":
			return "Xray-windows-arm64-v8a.zip", nil
		}
	}

	return "", fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
}

func unzipBinary(zipData []byte, dst string) error {
	r, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}

	target := binaryName()

	for _, f := range r.File {
		if filepath.Base(f.Name) != target {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("open file in zip: %w", err)
		}
		defer rc.Close()

		out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			return fmt.Errorf("create destination file: %w", err)
		}
		defer out.Close()

		if _, err := io.Copy(out, rc); err != nil {
			return fmt.Errorf("copy binary from zip: %w", err)
		}

		return nil
	}

	return fmt.Errorf("%s not found in archive", target)
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("create destination file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy file: %w", err)
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
