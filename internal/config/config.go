package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config 是应用配置，存放在可执行文件旁的 config.json。
type Config struct {
	ScanDir string `json:"scan_dir"` // 要扫描的 markdown 目录
	VSCode  string `json:"vscode"`   // VSCode Insider 可执行路径，默认 code-insiders
}

// Default 返回带默认值的配置。
func Default() Config {
	return Config{
		VSCode: "code-insiders",
	}
}

// Path 返回配置文件路径（可执行文件旁的 config.json）。
func Path() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(exe), "config.json"), nil
}

// Load 读取配置文件，不存在则返回默认配置。
func Load() (Config, error) {
	cfg := Default()
	p, err := Path()
	if err != nil {
		return cfg, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	if cfg.VSCode == "" {
		cfg.VSCode = "code-insiders"
	}
	return cfg, nil
}

// Save 写入配置文件。
func Save(cfg Config) error {
	p, err := Path()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}
