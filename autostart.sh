#!/usr/bin/env bash
# mdnowbar 开机自启动管理脚本（macOS launchd LaunchAgent）
# 用法: ./autostart.sh install|uninstall|start|stop|restart|status|logs
set -euo pipefail

LABEL="com.ikyrre.mdnowbar"
DOMAIN="gui/$(id -u)"
SERVICE="$DOMAIN/$LABEL"
REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY="$REPO_DIR/mdnowbar"
PLIST_DIR="$HOME/Library/LaunchAgents"
PLIST="$PLIST_DIR/$LABEL.plist"
LOG_OUT="$HOME/Library/Logs/mdnowbar.out.log"
LOG_ERR="$HOME/Library/Logs/mdnowbar.err.log"

usage() {
    cat <<EOF
用法: $0 <command>

  install    生成 LaunchAgent plist 并加载（开机自启 + 立即启动）
  uninstall  卸载并删除 plist（停止运行，取消开机自启）
  start      启动（若未安装则先安装）
  stop       停止运行（保留开机自启配置）
  restart    重启（重新编译后用这个）
  status     查看运行状态
  logs       查看日志（tail -f）
EOF
}

write_plist() {
    mkdir -p "$PLIST_DIR"
    cat > "$PLIST" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>$LABEL</string>
    <key>ProgramArguments</key>
    <array>
        <string>$BINARY</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <dict>
        <key>Crashed</key>
        <true/>
    </dict>
    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin</string>
    </dict>
    <key>StandardOutPath</key>
    <string>$LOG_OUT</string>
    <key>StandardErrorPath</key>
    <string>$LOG_ERR</string>
</dict>
</plist>
EOF
    plutil -lint "$PLIST" >/dev/null
}

is_loaded() {
    launchctl print "$SERVICE" >/dev/null 2>&1
}

cmd_install() {
    if [ ! -x "$BINARY" ]; then
        echo "未找到 $BINARY，尝试编译..."
        (cd "$REPO_DIR" && go build -o mdnowbar .)
    fi
    write_plist
    launchctl bootout "$SERVICE" 2>/dev/null || true
    launchctl bootstrap "$DOMAIN" "$PLIST"
    echo "已安装并启动：$BINARY"
    echo "日志: $LOG_OUT / $LOG_ERR"
}

cmd_uninstall() {
    launchctl bootout "$SERVICE" 2>/dev/null || true
    rm -f "$PLIST"
    echo "已卸载：$PLIST"
}

cmd_start() {
    if ! is_loaded; then
        if [ ! -f "$PLIST" ]; then
            cmd_install
            return
        fi
        launchctl bootstrap "$DOMAIN" "$PLIST"
    fi
    launchctl kickstart "$SERVICE" 2>/dev/null || true
    echo "已启动"
}

cmd_stop() {
    launchctl bootout "$SERVICE" 2>/dev/null || true
    echo "已停止（plist 保留，重新登录或 '$0 start' 可再次启动）"
}

cmd_restart() {
    if ! is_loaded; then
        echo "未加载，执行 install..."
        cmd_install
        return
    fi
    launchctl kickstart -k "$SERVICE"
    echo "已重启"
}

cmd_status() {
    if is_loaded; then
        launchctl print "$SERVICE" | sed -n '1,8p'
        echo "..."
        pgrep -fl "$BINARY" || true
    else
        echo "未安装/未加载"
        [ -f "$PLIST" ] && echo "plist 存在: $PLIST（用 '$0 start' 启动）"
    fi
}

cmd_logs() {
    touch "$LOG_OUT" "$LOG_ERR"
    tail -f "$LOG_OUT" "$LOG_ERR"
}

case "${1:-}" in
    install)   cmd_install ;;
    uninstall) cmd_uninstall ;;
    start)     cmd_start ;;
    stop)      cmd_stop ;;
    restart)   cmd_restart ;;
    status)    cmd_status ;;
    logs)      cmd_logs ;;
    *)         usage; exit 1 ;;
esac
