#!/system/bin/sh
# 限时读一次按键事件，最多等 $1 秒，防止 getevent 阻塞卡死安装
read_key_timeout() {
    local tmp="$MODPATH/.czero_key.tmp"
    getevent -qlc 1 >"$tmp" 2>/dev/null &
    local pid=$! n=0
    while [ "$n" -lt "$1" ] && kill -0 "$pid" 2>/dev/null; do
        sleep 1
        n=$((n + 1))
    done
    kill "$pid" 2>/dev/null
    cat "$tmp" 2>/dev/null
    rm -f "$tmp"
}

# 清空事件缓冲区，最多等 1 秒
clear_key_events() {
    read_key_timeout 1 >/dev/null
}

# 音量键检测，上键返回 0，下键返回 1，约 60 秒无按键返回 2
volume_key_detection() {
    clear_key_events
    local tries=0 key=""
    while [ "$tries" -lt 12 ]; do
        key=$(read_key_timeout 5 |
            awk '/KEY_VOLUMEUP/ {print "UP"; exit}
                 /KEY_VOLUMEDOWN/ {print "DOWN"; exit}')
        [ -n "$key" ] && break
        tries=$((tries + 1))
    done
    case "$key" in
        UP)   return 0 ;;
        DOWN) return 1 ;;
        *)    return 2 ;;
    esac
}

# 语言选择
choose_language() {
    ui_print " "
    ui_print "==================================="
    ui_print " Select Language / 选择语言 "
    ui_print "-----------------------------------"
    ui_print " [+] Volume Up   = English"
    ui_print " [-] Volume Down = 中文"
    ui_print "==================================="
    volume_key_detection
    case $? in
        0) UI_LANG=en ;;
        1) UI_LANG=zh ;;
        *) UI_LANG=zh; ui_print " (No key detected, default 中文)" ;;
    esac
}

# 显示选项菜单
show_menu() {
    ui_print " "
    ui_print "-----------------------------------"
    ui_print " $1 "
    [ -n "$2" ] && ui_print " $2 "
    ui_print "-----------------------------------"
    ui_print "1. [$L_AGREE] - $L_VOLUP"
    ui_print "2. [$L_CANCEL] - $L_VOLDOWN"
}

# 处理选择结果，超时视为取消
handle_choice() {
    volume_key_detection
    case $? in
        0) ui_print "$L_CHOSE$1"; return 0 ;;
        *) ui_print "$L_CHOSE$2"; return 1 ;;
    esac
}

# 定义可见文案
choose_language
if [ "$UI_LANG" = "en" ]; then
    L_AGREE="Agree";   L_CANCEL="Cancel"
    L_VOLUP="Volume Up"; L_VOLDOWN="Volume Down"
    L_CHOSE="You chose: "
    L_BANNER="CZero Installer"
    L_INHERIT_Q="Inherit previous config and black/white lists?"
    L_INHERIT_SUB="Not recommended if new features were added"
    L_NOCFG="Config not found; maybe the module was never installed"
    L_DONE="Installation complete"
    L_OPENURL="Opening project page:"
else
    L_AGREE="同意";   L_CANCEL="取消"
    L_VOLUP="音量键上"; L_VOLDOWN="音量键下"
    L_CHOSE="您选择了："
    L_BANNER="CZero 安装程序"
    L_INHERIT_Q="是否继承以往的配置和黑白名单？"
    L_INHERIT_SUB="若有添加新功能的情况下，不建议继承"
    L_NOCFG="未找到配置文件，可能是还未安装过模块"
    L_DONE="安装配置完成"
    L_OPENURL="正在打开项目主页："
fi

ui_print " "
ui_print "==================================="
ui_print "   $L_BANNER   "
ui_print "==================================="

# 继承配置
show_menu "$L_INHERIT_Q" "$L_INHERIT_SUB"
if handle_choice "$L_AGREE" "$L_CANCEL"; then
    if [ -f "/data/adb/modules/CZero/config.json" ]; then
        OLD="/data/adb/modules/CZero"
        for rel in \
            "list/Emptyfolder/directories.prop" \
            "list/Emptyfolder/emptyfolder_white.prop" \
            "list/clean_whitelist.prop" \
            "list/clean_paths.prop" \
            "list/app.json"; do
            [ -f "$OLD/$rel" ] && mv -f "$OLD/$rel" "$MODPATH/$rel"
        done
    else
        ui_print "$L_NOCFG"
    fi
fi

# 检测并移动basis目录
if [ -d "/data/adb/modules/CZero/basis" ]; then
    mv -f "/data/adb/modules/CZero/basis" "$MODPATH/"
fi

PROJECT_URL="https://github.com/Xocio/CZero"

ui_print " "
ui_print "==================================="
ui_print "   $L_DONE   "
ui_print "==================================="
ui_print " "
ui_print "$L_OPENURL"
ui_print "$PROJECT_URL"
sleep 5
am start -a android.intent.action.VIEW -d "$PROJECT_URL" >/dev/null 2>&1