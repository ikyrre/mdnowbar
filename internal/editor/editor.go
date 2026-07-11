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
