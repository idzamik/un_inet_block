package xray

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Lists struct {
	ConfigPath string
	DefaultURL string
	Client     *http.Client
}

func NewLists(configPath, defaultURL string) *Lists {
	return &Lists{
		ConfigPath: configPath,
		DefaultURL: defaultURL,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Lists) Install(source string) (string, error) {
	lines, err := c.loadAndParse(source)
	if err != nil {
		return "", err
	}

	if err := c.save(lines); err != nil {
		return "", err
	}

	return fmt.Sprintf("config installed: %d servers saved to %s", len(lines), c.ConfigPath), nil
}

func (c *Lists) Update(source string) (string, error) {
	lines, err := c.loadAndParse(source)
	if err != nil {
		return "", err
	}

	if err := c.save(lines); err != nil {
		return "", err
	}

	return fmt.Sprintf("config updated: %d servers saved to %s", len(lines), c.ConfigPath), nil
}

func (c *Lists) Delete() (string, error) {
	if _, err := os.Stat(c.ConfigPath); os.IsNotExist(err) {
		return "config is not installed", nil
	}

	if err := os.Remove(c.ConfigPath); err != nil {
		return "", fmt.Errorf("delete config: %w", err)
	}

	return fmt.Sprintf("config removed: %s", c.ConfigPath), nil
}

func (c *Lists) loadAndParse(source string) ([]string, error) {
	if source == "" {
		source = c.DefaultURL
	}

	raw, err := c.readSource(source)
	if err != nil {
		return nil, err
	}

	lines := parseServerList(raw)
	if len(lines) == 0 {
		return nil, fmt.Errorf("no valid vless links found in source")
	}

	return lines, nil
}

func (c *Lists) readSource(source string) ([]byte, error) {
	if isURL(source) {
		resp, err := c.Client.Get(source)
		if err != nil {
			return nil, fmt.Errorf("download source: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("download failed: %s", resp.Status)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read response body: %w", err)
		}

		return data, nil
	}

	data, err := os.ReadFile(source)
	if err != nil {
		return nil, fmt.Errorf("read local file: %w", err)
	}

	return data, nil
}

func parseServerList(data []byte) []string {
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	seen := make(map[string]struct{})
	var result []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.HasPrefix(strings.ToLower(line), "vless://") {
			continue
		}
		if _, ok := seen[line]; ok {
			continue
		}

		seen[line] = struct{}{}
		result = append(result, line)
	}

	return result
}

func (c *Lists) save(lines []string) error {
	dir := filepath.Dir(c.ConfigPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(c.ConfigPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

func isURL(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}
