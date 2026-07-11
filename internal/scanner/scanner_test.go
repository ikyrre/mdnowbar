package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseLine(t *testing.T) {
	cases := []struct {
		name   string
		line   string
		want   *Task
	}{
		{
			name: "未完成无now",
			line: "- [ ] 写文档",
			want: &Task{Text: "写文档", IsNow: false},
		},
		{
			name: "未完成带now",
			line: "- [ ] 写文档 #now",
			want: &Task{Text: "写文档", IsNow: true},
		},
		{
			name: "now在开头",
			line: "- [ ] #now 修复bug",
			want: &Task{Text: "修复bug", IsNow: true},
		},
		{
			name: "已完成x忽略",
			line: "- [x] 完成的任务",
			want: nil,
		},
		{
			name: "已完成X忽略",
			line: "- [X] 完成的任务",
			want: nil,
		},
		{
			name: "已取消忽略",
			line: "- [-] 取消的任务",
			want: nil,
		},
		{
			name: "缩进未完成",
			line: "  - [ ] 缩进任务 #now",
			want: &Task{Text: "缩进任务", IsNow: true},
		},
		{
			name: "星号列表标记",
			line: "* [ ] 星号任务",
			want: &Task{Text: "星号任务", IsNow: false},
		},
		{
			name: "加号列表标记",
			line: "+ [ ] 加号任务",
			want: &Task{Text: "加号任务", IsNow: false},
		},
		{
			name: "非todo行忽略",
			line: "这是普通段落",
			want: nil,
		},
		{
			name: "空括号无内容",
			line: "- [ ]  ",
			want: &Task{Text: "", IsNow: false},
		},
		{
			name: "now后缀无空格",
			line: "- [ ] 任务#now",
			want: &Task{Text: "任务", IsNow: true},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := parseLine(c.line, "/tmp/test.md", 1)
			if c.want == nil {
				if got != nil {
					t.Fatalf("期望 nil，得到 %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("期望 %+v，得到 nil", c.want)
			}
			if got.Text != c.want.Text {
				t.Errorf("Text: 期望 %q 得到 %q", c.want.Text, got.Text)
			}
			if got.IsNow != c.want.IsNow {
				t.Errorf("IsNow: 期望 %v 得到 %v", c.want.IsNow, got.IsNow)
			}
		})
	}
}

func TestScan(t *testing.T) {
	dir := t.TempDir()

	// 创建测试文件
	files := map[string]string{
		"a.md": "- [ ] 任务一 #now\n- [x] 已完成\n- [ ] 任务二\n- [-] 取消\n",
		"b.md": "- [ ] #now 紧急任务\n普通文本\n  - [ ] 缩进任务\n",
		"sub/c.md": "- [X] 完成\n- [ ] 子目录任务 #now\n",
		"readme.txt": "- [ ] 不是markdown\n",
	}
	for path, content := range files {
		full := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// .gitignore 忽略 sub 目录
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("sub/\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	// 期望未完成：a.md(任务一now, 任务二), b.md(紧急任务now, 缩进任务) = 4
	// sub 被 gitignore 忽略，readme.txt 非 md
	if res.Total != 4 {
		t.Errorf("Total: 期望 4 得到 %d", res.Total)
	}
	if res.Now != 2 {
		t.Errorf("Now: 期望 2 得到 %d", res.Now)
	}

	// 验证 sub 目录被跳过
	for _, task := range res.Tasks {
		if filepath.Base(task.File) == "c.md" {
			t.Errorf("sub/c.md 不应被扫描")
		}
	}
}

func TestScanHiddenDir(t *testing.T) {
	dir := t.TempDir()
	hidden := filepath.Join(dir, ".hidden", "x.md")
	if err := os.MkdirAll(filepath.Dir(hidden), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(hidden, []byte("- [ ] 隐藏任务 #now\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 0 {
		t.Errorf("隐藏目录不应被扫描，得到 %d", res.Total)
	}
}
