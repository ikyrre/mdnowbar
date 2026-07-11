package watcher

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// waitFor 轮询等待条件满足，超时则 t.Fatal。
func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestWatcherDebounceAndClose(t *testing.T) {
	dir := t.TempDir()
	var count atomic.Int32
	w, err := New(dir, func() { count.Add(1) })
	if err != nil {
		t.Fatal(err)
	}

	// 创建文件触发变化，等待防抖回调
	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte("- [ ] x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	waitFor(t, 2*time.Second, func() bool { return count.Load() >= 1 })
	if count.Load() == 0 {
		t.Fatal("期望防抖回调被触发")
	}

	// Close 可安全重复调用
	before := count.Load()
	w.Close()
	w.Close()

	// Close 后再触发变化，回调不应再增加
	if err := os.WriteFile(filepath.Join(dir, "b.md"), []byte("- [ ] y\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	time.Sleep(800 * time.Millisecond)
	if got := count.Load(); got != before {
		t.Errorf("Close 后不应再触发回调: before=%d after=%d", before, got)
	}
}
