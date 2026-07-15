# mdnowbar

[English](README.md) | [简体中文](README.zh-CN.md)

A macOS menu bar todo counter. It scans markdown files in a directory, counts unfinished todo items, and shows the number of urgent `#now`-tagged tasks in the menu bar.

## Features

- Recursively scans `.md` files (respects `.gitignore`, skips hidden directories)
- Counts unfinished todos (`- [ ]`), excluding completed (`- [x]`) and cancelled (`- [-]`)
- Recognizes `#now`-tagged urgent tasks
- Menu bar icon shows the number of `#now` tasks
- Click the icon to see `now/total` and the full list of now tasks
- Click a task to jump to its file and line in VSCode, or open it in Obsidian
- Real-time updates on file changes (fsnotify + debounce)

## Usage

```bash
# Build
go build -o mdnowbar .

# Run
./mdnowbar
```

On first run, click the menu bar icon →「设置扫描目录…」(Set scan directory) to pick the markdown directory to scan.

## Markdown format

```markdown
- [ ] A normal unfinished task
- [ ] An urgent task #now
- [x] A completed task
- [-] A cancelled task
```

## Configuration

The config file `config.json` lives next to the executable:

```json
{
  "scan_dir": "/path/to/your/notes",
  "vscode": "code",
  "editor": "vscode",
  "obsidian_vault": "",
  "obsidian_advanced_uri": false
}
```

- `vscode` — VSCode executable name (e.g. `code`, `code-insiders`). If PATH can't find it when launched from the GUI, set the full path here.
- `editor` — which app opens a task on click: `vscode` (default) or `obsidian`.
- `obsidian_vault` — Obsidian vault name. Leave empty to auto-resolve from `~/Library/Application Support/obsidian/obsidian.json` by matching `scan_dir`.
- `obsidian_advanced_uri` — use the `obsidian://advanced-uri` scheme to jump to a specific **line**. Requires the [Advanced URI](https://github.com/Vinzent03/obsidian-advanced-uri) plugin. When `false`, the native `obsidian://open` is used, which opens the file but cannot position the cursor on a line.

## Project structure

```
mdnowbar/
├── main.go                    # Entry point
├── internal/
│   ├── scanner/               # Markdown scanning & parsing
│   ├── config/                # Config read/write
│   ├── editor/                # VSCode jump
│   ├── watcher/               # File watching
│   └── tray/                  # Menu bar UI
└── assets/
    └── icon.png               # Menu bar icon
```

## License

[MIT License](LICENSE), Copyright (c) 2026 ikyrre
