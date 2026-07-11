package main

import (
	"log"

	"mytodo/internal/config"
	"mytodo/internal/scanner"
	"mytodo/internal/tray"
	"mytodo/internal/watcher"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("加载配置失败: %v，使用默认配置", err)
	}

	app := tray.New(cfg, nil)

	// rescan 执行一次扫描并更新菜单
	rescan := func() {
		dir := app.ScanDir()
		if dir == "" {
			app.UpdateResult(&scanner.Result{})
			return
		}
		r, err := scanner.Scan(dir)
		if err != nil {
			log.Printf("扫描失败: %v", err)
			return
		}
		app.UpdateResult(r)
	}

	// 设置 tray 的 onRescan 回调
	app.SetOnRescan(rescan)

	// 启动文件监听（如果已配置目录）
	var w *watcher.Watcher
	startWatcher := func() {
		if w != nil {
			w.Close()
		}
		dir := app.ScanDir()
		if dir == "" {
			return
		}
		var err error
		w, err = watcher.New(dir, rescan)
		if err != nil {
			log.Printf("启动监听失败: %v", err)
		}
	}

	// 初始扫描
	rescan()
	startWatcher()

	// 监听目录变化后重启 watcher
	app.SetOnDirChange(startWatcher)

	// 启动状态栏（阻塞）
	app.Run()

	if w != nil {
		w.Close()
	}
}
