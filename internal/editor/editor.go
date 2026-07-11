package editor

import (
	"fmt"
	"os/exec"
)

// OpenAt 用 VSCode Insider 打开 file 的第 line 行。
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
