# mdnowbar

[English](README.md) | [简体中文](README.zh-CN.md)

A macOS menu bar todo counter. It scans markdown files in a directory, counts unfinished todo items, and shows the number of urgent `#now`-tagged tasks in the menu bar.

## Features

- Recursively scans `.md` files (respects `.gitignore`, skips hidden directories)
- Counts unfinished todos (`- [ ]`), excluding completed (`- [x]`) and cancelled (`- [-]`)
- Recognizes `#now`-tagged urgent tasks
- Menu bar icon shows the number of `#now` tasks
- Click the icon to see `now/total` and the full list of now tasks
- Click a task to jump to its file and line in VSCode
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
  "vscode": "code"
}
```

The `vscode` field is the VSCode executable name (e.g. `code`, `code-insiders`). If PATH can't find it when launched from the GUI, set the full path here.

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
