package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
)

// Task 表示一个未完成的 markdown todo 项。
type Task struct {
	File  string // 绝对路径
	Line  int    // 行号（从 1 开始）
	Text  string // 任务文本（去掉标记和 #now 后的纯文本）
	IsNow bool   // 是否包含 #now 标记
}

// Result 是一次扫描的结果。
type Result struct {
	Total int    // 未完成任务总数（含 now）
	Now   int    // #now 任务数
	Tasks []Task // 全部未完成任务
}

// todoRe 匹配 markdown todo 行：`- [ ]`、`- [x]`、`- [-]`
// 支持前导空白（缩进）和 `-` 或 `*` 作为列表标记。
var todoRe = regexp.MustCompile(`^\s*[-*+]\s+\[([ xX-])\]\s+(.*)$`)

// nowTag 是紧急标记。
const nowTag = "#now"

// Scan 递归扫描 dir 下的 .md 文件，返回未完成 todo 列表。
// 会读取 dir 根目录的 .gitignore（若存在）进行过滤。
func Scan(dir string) (*Result, error) {
	dir = filepath.Clean(dir)
	var gi *ignore.GitIgnore
	if fi, err := os.Stat(filepath.Join(dir, ".gitignore")); err == nil && !fi.IsDir() {
		gi, err = ignore.CompileIgnoreFile(filepath.Join(dir, ".gitignore"))
		if err != nil {
			gi = nil // 解析失败则忽略，不阻断扫描
		}
	}

	res := &Result{}
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, path)

		// gitignore 过滤
		if gi != nil && rel != "." {
			if gi.MatchesPath(rel) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// 跳过隐藏目录（.git 等）
		if d.IsDir() {
			if path != dir && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// 只处理 .md 文件
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}

		tasks, err := scanFile(path)
		if err != nil {
			return err
		}
		res.Tasks = append(res.Tasks, tasks...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	for _, t := range res.Tasks {
		res.Total++
		if t.IsNow {
			res.Now++
		}
	}
	return res, nil
}

// scanFile 解析单个 markdown 文件，返回其中的未完成 todo。
func scanFile(path string) ([]Task, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var tasks []Task
	sc := bufio.NewScanner(f)
	// 支持较长行
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lineNum := 0
	for sc.Scan() {
		lineNum++
		t := parseLine(sc.Text(), path, lineNum)
		if t != nil {
			tasks = append(tasks, *t)
		}
	}
	return tasks, sc.Err()
}

// parseLine 解析单行，若是未完成 todo 则返回 Task，否则返回 nil。
//
//	[ ] => 未完成
//	[x] / [X] => 完成，忽略
//	[-] => 取消，忽略
func parseLine(line, file string, lineNum int) *Task {
	m := todoRe.FindStringSubmatch(line)
	if m == nil {
		return nil
	}
	status := m[1]
	if status == "x" || status == "X" || status == "-" {
		return nil
	}
	// status == " " => 未完成
	text := strings.TrimSpace(m[2])
	isNow := false
	if strings.Contains(text, nowTag) {
		isNow = true
		text = strings.TrimSpace(strings.ReplaceAll(text, nowTag, ""))
	}
	return &Task{
		File:  file,
		Line:  lineNum,
		Text:  text,
		IsNow: isNow,
	}
}
