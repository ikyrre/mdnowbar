package tray

import (
	_ "embed"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ikyrre/mdnowbar/internal/config"
	"github.com/ikyrre/mdnowbar/internal/editor"
	"github.com/ikyrre/mdnowbar/internal/scanner"

	"fyne.io/systray"
)

//go:embed icon.png
var iconData []byte

// App 是状态栏应用控制器。
type App struct {
	cfg         config.Config
	mu          sync.Mutex
	menuMu      sync.Mutex // 保护 refreshMenu 的菜单重建序列，避免并发 Reset/Add 交叉
	result      *scanner.Result
	onRescan    func() // 触发重新扫描的回调
	onDirChange func() // 扫描目录变化时的回调
}

// New 创建 App 实例。onRescan 用于触发重新扫描。
func New(cfg config.Config, onRescan func()) *App {
	return &App{
		cfg:      cfg,
		onRescan: onRescan,
	}
}

// SetOnRescan 设置触发重新扫描的回调。
func (a *App) SetOnRescan(fn func()) {
	a.mu.Lock()
	a.onRescan = fn
	a.mu.Unlock()
}

// SetOnDirChange 设置扫描目录变化时的回调（用于重启 watcher）。
func (a *App) SetOnDirChange(fn func()) {
	a.mu.Lock()
	a.onDirChange = fn
	a.mu.Unlock()
}

// UpdateResult 在扫描完成后更新结果并刷新菜单。
func (a *App) UpdateResult(r *scanner.Result) {
	a.mu.Lock()
	a.result = r
	a.mu.Unlock()
	a.refreshMenu()
}

// Run 启动 systray（阻塞主线程）。
func (a *App) Run() {
	systray.Run(a.onReady, a.onExit)
}

func (a *App) onReady() {
	systray.SetTemplateIcon(iconData, iconData)
	systray.SetTitle("")
	systray.SetTooltip("mdnowbar")
	a.refreshMenu()
}

func (a *App) onExit() {
	// 清理
}

// refreshMenu 重建整个菜单。
// 持 menuMu 串行化整个 Reset+Add 序列，避免多个 goroutine 并发重建导致菜单错乱。
func (a *App) refreshMenu() {
	a.menuMu.Lock()
	defer a.menuMu.Unlock()

	systray.ResetMenu()

	a.mu.Lock()
	r := a.result
	a.mu.Unlock()

	if r == nil {
		systray.SetTitle("…")
		mLoading := systray.AddMenuItem("加载中…", "")
		mLoading.Disable()
	} else {
		systray.SetTitle(fmt.Sprintf("%d", r.Now))
		systray.SetTooltip(fmt.Sprintf("%d now / %d total", r.Now, r.Total))

		// 标题行
		mTitle := systray.AddMenuItem(fmt.Sprintf("%d now / %d total", r.Now, r.Total), "")
		mTitle.Disable()

		// 列出 now 任务
		nowTasks := a.nowTasks(r)
		if len(nowTasks) > 0 {
			systray.AddSeparator()
			for _, t := range nowTasks {
				a.addTaskItem(t)
			}
		}
	}

	systray.AddSeparator()

	// 设置扫描目录
	mSetDir := systray.AddMenuItem("设置扫描目录…", "选择要扫描的 markdown 目录")
	go a.handleSetDir(mSetDir)

	// 手动刷新
	mRefresh := systray.AddMenuItem("刷新", "重新扫描")
	go a.handleRefresh(mRefresh)

	systray.AddSeparator()

	// 退出
	mQuit := systray.AddMenuItem("退出", "退出 mdnowbar")
	go func() {
		for range mQuit.ClickedCh {
			systray.Quit()
		}
	}()
}

// nowTasks 返回所有带 #now 标记的任务。
func (a *App) nowTasks(r *scanner.Result) []scanner.Task {
	var tasks []scanner.Task
	for _, t := range r.Tasks {
		if t.IsNow {
			tasks = append(tasks, t)
		}
	}
	return tasks
}

// addTaskItem 添加一个任务菜单项并绑定点击事件。
func (a *App) addTaskItem(t scanner.Task) {
	label := formatTaskLabel(t)
	item := systray.AddMenuItem(label, fmt.Sprintf("%s:%d", t.File, t.Line))
	go func() {
		for range item.ClickedCh {
			if err := editor.OpenAt(a.cfg.VSCode, t.File, t.Line); err != nil {
				log.Printf("打开编辑器失败 (%s:%d): %v", t.File, t.Line, err)
			}
		}
	}()
}

// formatTaskLabel 格式化任务显示文本：文件名:行号 - 任务文本。
func formatTaskLabel(t scanner.Task) string {
	name := filepath.Base(t.File)
	text := t.Text
	// 截断过长文本
	if len([]rune(text)) > 40 {
		text = string([]rune(text)[:37]) + "…"
	}
	return fmt.Sprintf("%s:%d - %s", name, t.Line, text)
}

// handleSetDir 处理设置扫描目录。
func (a *App) handleSetDir(item *systray.MenuItem) {
	for range item.ClickedCh {
		dir, err := pickFolder()
		if err != nil || dir == "" {
			continue
		}
		a.mu.Lock()
		a.cfg.ScanDir = dir
		onDirChange := a.onDirChange
		onRescan := a.onRescan
		a.mu.Unlock()
		if err := config.Save(a.cfg); err != nil {
			log.Printf("保存配置失败: %v", err)
		}
		if onDirChange != nil {
			onDirChange()
		}
		if onRescan != nil {
			onRescan()
		}
	}
}

// handleRefresh 处理手动刷新。
func (a *App) handleRefresh(item *systray.MenuItem) {
	for range item.ClickedCh {
		a.mu.Lock()
		onRescan := a.onRescan
		a.mu.Unlock()
		if onRescan != nil {
			onRescan()
		}
	}
}

// pickFolder 用 macOS osascript 弹出文件夹选择对话框。
func pickFolder() (string, error) {
	script := `POSIX path of (choose folder with prompt "选择要扫描的 markdown 目录")`
	cmd := exec.Command("osascript", "-e", script)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// ScanDir 返回当前配置的扫描目录。
func (a *App) ScanDir() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.cfg.ScanDir
}
