package watcher

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher 递归监听目录变化，防抖触发回调。
type Watcher struct {
	w       *fsnotify.Watcher
	root    string
	onChange func()
	stop    chan struct{}
	wg      sync.WaitGroup
	mu      sync.Mutex
	timer   *time.Timer
	debounce time.Duration
}

// New 创建一个递归监听器。onChange 在文件变化（防抖后）被调用。
func New(root string, onChange func()) (*Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w := &Watcher{
		w:        fw,
		root:     filepath.Clean(root),
		onChange:  onChange,
		stop:      make(chan struct{}),
		debounce:  300 * time.Millisecond,
	}
	if err := w.addRecursive(w.root); err != nil {
		fw.Close()
		return nil, err
	}
	w.wg.Add(1)
	go w.loop()
	return w, nil
}

// addRecursive 递归添加目录及其子目录到监听器。
func (w *Watcher) addRecursive(dir string) error {
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // 忽略错误继续
		}
		if !d.IsDir() {
			return nil
		}
		// 跳过隐藏目录
		if path != w.root && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}
		return w.w.Add(path)
	})
}

// loop 处理事件循环。
func (w *Watcher) loop() {
	defer w.wg.Done()
	for {
		select {
		case <-w.stop:
			return
		case ev, ok := <-w.w.Events:
			if !ok {
				return
			}
			// 新建目录时动态添加监听
			if ev.Has(fsnotify.Create) {
				if fi, err := os.Stat(ev.Name); err == nil && fi.IsDir() {
					_ = w.addRecursive(ev.Name)
				}
			}
			w.schedule()
		case _, ok := <-w.w.Errors:
			if !ok {
				return
			}
		}
	}
}

// schedule 防抖触发 onChange。
func (w *Watcher) schedule() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.timer != nil {
		w.timer.Stop()
	}
	w.timer = time.AfterFunc(w.debounce, func() {
		if w.onChange != nil {
			w.onChange()
		}
	})
}

// Close 停止监听并释放资源。
func (w *Watcher) Close() {
	close(w.stop)
	_ = w.w.Close()
	w.wg.Wait()
}
