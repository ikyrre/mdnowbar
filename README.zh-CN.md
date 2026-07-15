# mdnowbar

[English](README.md) | [简体中文](README.zh-CN.md)

Mac 状态栏 todo 计数器。扫描指定目录下的 markdown 文件，统计未完成 todo 项数量，在状态栏显示带 `#now` 标记的紧急任务数。

## 功能

- 递归扫描指定目录的 `.md` 文件（尊重 `.gitignore`，跳过隐藏目录）
- 统计未完成 todo 项（`- [ ]`），排除已完成（`- [x]`）和已取消（`- [-]`）
- 识别 `#now` 标记的紧急任务
- 状态栏图标显示 `#now` 任务数量
- 点击图标显示 `now/total` 及全部 now 任务列表
- 点击任务项用 VSCode 跳转到对应文件和行，或用 Obsidian 打开
- 文件变化实时更新（fsnotify + 防抖）

## 使用

```bash
# 编译
go build -o mdnowbar .

# 运行
./mdnowbar
```

首次运行后，点击状态栏图标 ->「设置扫描目录…」选择要扫描的 markdown 目录。

## markdown 格式

```markdown
- [ ] 普通未完成任务
- [ ] 紧急任务 #now
- [x] 已完成任务
- [-] 已取消任务
```

## 配置

配置文件 `config.json` 位于可执行文件旁：

```json
{
  "scan_dir": "/path/to/your/notes",
  "vscode": "code",
  "editor": "vscode",
  "obsidian_vault": "",
  "obsidian_advanced_uri": false
}
```

- `vscode` - VSCode 可执行文件名（如 `code`、`code-insiders`），GUI 启动时若 PATH 找不到，可在配置中设置完整路径。
- `editor` - 点击任务时用哪个应用打开：`vscode`（默认）或 `obsidian`。
- `obsidian_vault` - Obsidian vault 名。留空则根据 `scan_dir` 自动从 `~/Library/Application Support/obsidian/obsidian.json` 解析。
- `obsidian_advanced_uri` - 是否用 `obsidian://advanced-uri` scheme 以跳转到具体**行号**。需安装 [Advanced URI](https://github.com/Vinzent03/obsidian-advanced-uri) 插件。为 `false` 时使用原生 `obsidian://open`，只能打开文件、无法定位行。

## 项目结构

```
mdnowbar/
├── main.go                    # 入口
├── internal/
│   ├── scanner/               # markdown 扫描与解析
│   ├── config/                # 配置读写
│   ├── editor/                # VSCode 跳转
│   ├── watcher/               # 文件监听
│   └── tray/                  # 状态栏 UI
└── assets/
    └── icon.png               # 状态栏图标
```

## License

[MIT License](LICENSE), Copyright (c) 2026 ikyrre
