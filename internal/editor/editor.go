package editor

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// OpenAt 用 VSCode 打开 file 的第 line 行。
// vscodePath 为可执行路径（如 code-insiders 或完整路径）。
func OpenAt(vscodePath, file string, line int) error {
	target := fmt.Sprintf("%s:%d", file, line)
	cmd := exec.Command(vscodePath, "--goto", target)
	return cmd.Start()
}

// Lookup 校验 vscodePath 在 PATH 中是否可找到。
// 用于启动时提示用户配置正确路径（GUI 启动时 PATH 可能不含 /usr/local/bin）。
func Lookup(vscodePath string) error {
	_, err := exec.LookPath(vscodePath)
	return err
}

// OpenOptions 描述如何打开一个文件位置。
type OpenOptions struct {
	Editor              string // "vscode"（默认）或 "obsidian"
	VSCodePath          string // VSCode 可执行路径
	ScanDir             string // 扫描目录，用于自动解析 obsidian vault
	ObsidianVault       string // 手动指定 vault 名，留空则自动解析
	ObsidianAdvancedURI bool   // 是否使用 advanced-uri scheme（支持行号，需插件）
	File                string // 目标文件绝对路径
	Line                int    // 行号（从 1 开始）
}

// Open 根据 opts.Editor 分发到 VSCode 或 Obsidian。
func Open(opts OpenOptions) error {
	switch strings.ToLower(strings.TrimSpace(opts.Editor)) {
	case "obsidian":
		return openObsidian(opts)
	default:
		return OpenAt(opts.VSCodePath, opts.File, opts.Line)
	}
}

// openObsidian 通过 obsidian:// URL scheme 打开文件。
func openObsidian(opts OpenOptions) error {
	vaultName, vaultPath, err := ResolveVault(opts.ScanDir, opts.ObsidianVault)
	if err != nil {
		return fmt.Errorf("解析 obsidian vault 失败: %w", err)
	}

	rel, err := filepath.Rel(vaultPath, opts.File)
	if err != nil {
		return fmt.Errorf("计算相对路径失败: %w", err)
	}
	rel = strings.TrimSuffix(rel, ".md")
	rel = filepath.ToSlash(rel)

	var u string
	if opts.ObsidianAdvancedURI {
		// 需要 Obsidian Advanced URI 插件；filepath 带扩展名，line 支持行号。
		u = fmt.Sprintf("obsidian://advanced-uri?vault=%s&filepath=%s&line=%d",
			uriEscape(vaultName), uriEscape(rel+".md"), opts.Line)
	} else {
		// 原生 obsidian://open 不支持行号，仅打开文件。
		u = fmt.Sprintf("obsidian://open?vault=%s&file=%s",
			uriEscape(vaultName), uriEscape(rel))
	}
	return exec.Command("open", u).Start()
}

// uriEscape 对 obsidian:// 参数值做 URL 编码。
// 用 %20 编码空格而非 +（Obsidian 的 URI 处理器不会把 + 解码为空格）。
func uriEscape(s string) string {
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}

// obsidianVaultsFile 返回 obsidian.json（vault 注册表）路径。
func obsidianVaultsFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Application Support", "obsidian", "obsidian.json"), nil
}

type obsidianConfig struct {
	Vaults map[string]obsidianVault `json:"vaults"`
}

type obsidianVault struct {
	Path string `json:"path"`
	Open bool   `json:"open"`
}

// ResolveVault 解析 scanDir 所属的 obsidian vault，返回 vault 名（目录名）与绝对路径。
// 若 override 非空，则优先匹配目录名等于 override 的 vault；
// 否则取路径最深（最具体）的包含 scanDir 的 vault。
func ResolveVault(scanDir, override string) (name, vaultPath string, err error) {
	p, err := obsidianVaultsFile()
	if err != nil {
		return "", "", fmt.Errorf("定位 obsidian.json 失败: %w", err)
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return "", "", fmt.Errorf("读取 obsidian.json 失败: %w（请先打开一次 Obsidian）", err)
	}
	var cfg obsidianConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", "", fmt.Errorf("解析 obsidian.json 失败: %w", err)
	}

	scanDir = filepath.Clean(scanDir)
	var best string
	for _, v := range cfg.Vaults {
		if v.Path == "" {
			continue
		}
		vp := filepath.Clean(v.Path)
		base := filepath.Base(vp)
		if override != "" && base == override {
			return base, vp, nil
		}
		rel, rerr := filepath.Rel(vp, scanDir)
		if rerr != nil {
			continue
		}
		// rel 不以 ".." 开头 => scanDir 位于 vp 内（含相等）
		if !strings.HasPrefix(rel, "..") {
			if len(vp) > len(best) {
				best = vp
			}
		}
	}
	if best != "" {
		return filepath.Base(best), best, nil
	}
	if override != "" {
		return "", "", fmt.Errorf("未找到名为 %q 的 obsidian vault", override)
	}
	return "", "", fmt.Errorf("未找到包含 %q 的 obsidian vault", scanDir)
}
