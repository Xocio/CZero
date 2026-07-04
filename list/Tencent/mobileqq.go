package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	configPath     = "/data/adb/modules/CZero/config.json"
	logDir         = "/data/adb/modules/CZero/log"
	logTag         = "[QQ] "
	statisticsFile = "/data/adb/modules/CZero/basis/basis.prop"
	basisLockFile  = "/data/adb/modules/CZero/basis/.basis.lock"
	whitelistFile  = "/data/adb/modules/CZero/list/clean_whitelist.prop"

	statsKey = "qq" // 统计键名
)

type CleanRule struct {
	Name       string
	Paths      []string
	Regexes    []string
	Exclusions []string
}

var (
	whitelistPaths  []string // 白名单路径列表
	timeBarrierDays int      // 时序屏障天数
)

// JSON 配置
type Config struct {
	General struct {
		Log                 bool `json:"log"`
		TemporalBarrierDays int  `json:"temporal_barrier_days"`
	} `json:"general"`
	AppClean struct {
		Qq struct {
			Enabled  bool `json:"enabled"`
			Enhanced bool `json:"enhanced"`
		} `json:"qq"`
	} `json:"app_clean"`
}

// 配置
var cfg Config

func loadConfig() {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &cfg)
}

// 时序屏障，读取保护天数
func loadTimeBarrier() {
	timeBarrierDays = cfg.General.TemporalBarrierDays
}

// 检查文件是否超过时序屏障时间
func isOlderThanBarrier(filePath string, fileInfo fs.DirEntry) bool {
	if timeBarrierDays <= 0 {
		return true // 屏障为0或负数时允许清理所有文件
	}
	info, err := fileInfo.Info()
	if err != nil {
		writeLog(fmt.Sprintf("[时间检查错误] %s: %v\n", filePath, err))
		return false // 获取不到时间信息时保护文件
	}
	barrierTime := time.Now().AddDate(0, 0, -timeBarrierDays)
	isOlder := info.ModTime().Before(barrierTime)
	if !isOlder && logEnabled {
		writeLog(fmt.Sprintf("[时序保护] 跳过文件: %s (修改时间: %s, 屏障: %d天)\n",
			filePath, info.ModTime().Format("2006-01-02 15:04:05"), timeBarrierDays))
	}
	return isOlder
}

// 白名单，启动时一次性加载
func loadWhitelist() {
	whitelistPaths = nil
	data, err := os.ReadFile(whitelistFile)
	if err != nil {
		if !os.IsNotExist(err) {
			writeLog(fmt.Sprintf("[白名单错误] 无法读取白名单文件: %v\n", err))
		}
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			whitelistPaths = append(whitelistPaths, line)
			writeLog(fmt.Sprintf("[白名单] 已加载: %s\n", line))
		}
	}
	writeLog(fmt.Sprintf("[白名单] 共加载 %d 个路径\n", len(whitelistPaths)))
}

func isWhitelisted(path string) bool {
	normalized := filepath.ToSlash(path)
	for _, whitePath := range whitelistPaths {
		if strings.Contains(normalized, whitePath) {
			if logEnabled {
				writeLog(fmt.Sprintf("[白名单] 跳过路径: %s (匹配: %s)\n", path, whitePath))
			}
			return true
		}
	}
	return false
}

var (
	normalRules = []CleanRule{
		{
			Name: "QQ基础缓存",
			Paths: []string{
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/tencent/mobileqq/.apollo/dress",
				"/storage/emulated/%s/Tencent/xglog",
				"/storage/emulated/%s/Tencent/wtlogin",
				"/storage/emulated/%s/Tencent/wns",
				"/storage/emulated/%s/Tencent/Tim/RedPacket",
				"/storage/emulated/%s/Tencent/Tim/QQmail",
				"/storage/emulated/%s/Tencent/Tim/log",
				"/storage/emulated/%s/Tencent/QQmail/imagecache",
				"/storage/emulated/%s/Tencent/QQmail/tmp",
				"/storage/emulated/%s/Tencent/QQmail/qmlog",
				"/storage/emulated/%s/Tencent/QQmail/nickIcon",
				"/storage/emulated/%s/Tencent/QQLite/log",
				"/storage/emulated/%s/Tencent/QQLite/diskcache",
				"/storage/emulated/%s/Tencent/QQLite/ArkApp/Crash",
				"/storage/emulated/%s/Tencent/QQLite/ArkApp/Cache",
				"/storage/emulated/%s/Tencent/QQfile_recv/.tmp",
				"/storage/emulated/%s/Tencent/QQ_Collection",
				"/storage/emulated/%s/Tencent/msflogs",
				"/storage/emulated/%s/Tencent/MobileQQ/thumb2",
				"/storage/emulated/%s/Tencent/MobileQQ/thumb",
				"/storage/emulated/%s/Tencent/MobileQQ/RedPacket",
				"/storage/emulated/%s/Tencent/MobileQQ/ocr/cache",
				"/storage/emulated/%s/Tencent/MobileQQ/log",
				"/storage/emulated/%s/Tencent/MobileQQ/avatarPendantIcons",
				"/storage/emulated/%s/Tencent/MobileQQ/audioCache",
				"/storage/emulated/%s/Tencent/Midas/Log",
				"/storage/emulated/%s/Tencent/imsdklogs",
				"/storage/emulated/%s/Tencent/beacon",
				"/storage/emulated/%s/Tencent/ams",
				"/storage/emulated/%s/qqstory/.tmp/.tmp",
				"/storage/emulated/%s/Android/data/com.tencent.qqlite/qzone",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/files/QQShowDownload",
				"/storage/emulated/%s/Android/data/com.tencent.qqlite/files/tencent/msflogs",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/files/zootopia_download",
				"/storage/emulated/%s/Android/data/com.tencent.qqlite/files/tencent/MobileQQ/head/_hd",
				"/storage/emulated/%s/Android/data/com.tencent.qqlite/cache",
				"/storage/emulated/%s/Android/data/com.tencent.tim/Tencent/Tim/head",
				"/storage/emulated/%s/Android/data/com.tencent.tim/Tencent/Tim/diskcache",
				"/storage/emulated/%s/Android/data/com.tencent.tim/Tencent/Tim/chatpic/chatthumb",
				"/storage/emulated/%s/Android/data/com.tencent.tim/Tencent/QQ_Collection",
				"/storage/emulated/%s/Android/data/com.tencent.tim/Tencent/mini",
				"/storage/emulated/%s/Android/data/com.tencent.tim/qzone/gallerytmp",
				"/storage/emulated/%s/Android/data/com.tencent.tim/qzone/zip_cache",
				"/storage/emulated/%s/Android/data/com.tencent.tim/qzone/video_cache",
				"/storage/emulated/%s/Android/data/com.tencent.tim/files/tencent/MobileQQ/head",
				"/storage/emulated/%s/Android/data/com.tencent.tim/files/tencent/tbs_live_log",
				"/storage/emulated/%s/Android/data/com.tencent.tim/files/tbslog",
				"/storage/emulated/%s/Android/data/com.tencent.tim/cache",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/QQfile_recv/.tmp",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/QQ_Collection",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/msflogs",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/thumb",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/head",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/DoutuRes",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/diskcache",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/.emotionsm",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/.vipicon",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/.gift",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/mini",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/qzone/zip_cache",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/qzone/video_cache",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/qzone/rapid_comment",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/qzone/gallerytmp",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/qzone/audio",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/qcircle",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/files/VideoCache",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/files/tencent/tbs_live_log",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/files/tencent/tbs",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/files/tencent/msflogs",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/files/tencent/MobileQQ/head",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/files/tencent/com/tencent/mobileqq/avsdk/log",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/files/tbslog",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/files/tbs",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/files/ShadowPlugin_Base/Tencent/now/skin",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/files/ShadowPlugin_Base/Tencent/now/log",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/files/ShadowPlugin_Base/Tencent/now/crash",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/files/QWallet",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/files/onelog",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/files/now",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/files/data",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/files/.info",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/cache",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/Scribble",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/files/QWallet/.preloaduni",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/QQ_Images/QQEditPic",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/qbosssplahAD",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/.pendant",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/.font_info",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/files/ae",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/qzone",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/files/ShadowPlugin_RoomBiz",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/wxminiapp",
				"/sdcard/Tencent/cache",
				"/sdcard/Tencent/MobileQQ/diskcache",
				"/sdcard/Tencent/MobileQQ/Scribble",
				"/sdcard/Tencent/MobileQQ/ScribbleCache",
				"/sdcard/Tencent/MobileQQ/qav",
				"/sdcard/Tencent/MobileQQ/qqmusic",
				"/sdcard/Tencent/MobileQQ/pddata",
				"/storage/emulated/%s/tencent/QQGallery/log",
				"/sdcard/Tencent/MobileQQ/photo",
				"/sdcard/Tencent/MobileQQ/chatpic",
				"/sdcard/Tencent/MobileQQ/thumb",
				"/sdcard/Tencent/MobileQQ/QQ_Images",
				"/sdcard/Tencent/MobileQQ/QQEditPic",
				"/sdcard/Tencent/MobileQQ/hotpic",
				"/sdcard/Tencent/MobileQQ/shortvideo",
				"/sdcard/Tencent/MobileQQ/qbosssplahAD",
				"/sdcard/Tencent/MobileQQ/.apollo",
				"/sdcard/Tencent/MobileQQ/vas",
				"/sdcard/Tencent/MobileQQ/lottie",
				"/sdcard/Tencent/mini",
				"/sdcard/Tencent/TMAssistantSDK",
				"/sdcard/Tencent/.font_info",
				"/sdcard/Tencent/.hiboom_font",
				"/sdcard/Tencent/.gift",
				"/sdcard/Tencent/.trooprm/enter_effects",
				"/sdcard/Tencent/tbs",
				"/sdcard/Tencent/.pendant",
				"/sdcard/Tencent/.profilecard",
				"/sdcard/Tencent/.sticker_recommended_pics",
				"/sdcard/Tencent/pe",
				"/sdcard/Tencent/.emotionsm",
				"/sdcard/Tencent/.vaspoke",
				"/sdcard/Tencent/newpoke",
				"/sdcard/Tencent/poke",
				"/sdcard/Tencent/.vipicon",
				"/sdcard/Tencent/DoutuRes",
				"/sdcard/Tencent/funcall",
				"/sdcard/Tencent/QQfile_recv/.trooptmp",
				"/sdcard/Tencent/QQfile_recv/.tmp",
				"/sdcard/Tencent/QQfile_recv/.thumbnails",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/diskcache",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/Scribble",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/ScribbleCache",
				"/storage/emulated/%s/tencent/QQSecure/Athena",
				"/storage/emulated/%s/tencent/tmp/*.log",
				"/storage/emulated/%s/tencent/MobileQQ/log",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/qav",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/qqmusic",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/pddata",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/cache",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/photo",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/chatpic",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/thumb",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/QQ_Images",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/QQEditPic",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/hotpic",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/shortvideo",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/qbosssplahAD",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/.apollo",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/vasrm",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/lottie",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/.hiboom_font",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/MobileQQ/.gift",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/.trooprm/enter_effects",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/head",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/.pendant",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/.profilecard",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/.sticker_recommended_pics",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/pe",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/.vaspoke",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/newpoke",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/poke",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/.vipicon",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/DoutuRes",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/funcall",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/QQfile_recv/.trooptmp",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/QQfile_recv/.tmp",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/QQfile_recv/.thumbnails",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/qcircle",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/files/.info",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/files/onelog",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/files/ae/playshow",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/files/tencent/msflogs",
				"/storage/emulated/%s/tencent/msflogs",
				"/storage/emulated/%s/tencent/wns/Logs",
				"/storage/emulated/%s/tencent/wtlogin",
				"/storage/emulated/%s/tencent/QQmail/log",
				"/storage/emulated/%s/tencent/QQmail/qmlog",
				"/storage/emulated/%s/tencent/qqimsecure/pt",
				"/storage/emulated/%s/tencent/MobileQQ/ocr/cache",
				"/sdcard/Android/data/com.tencent.mobileqq/files/.info",
				"/sdcard/Android/data/com.tencent.mobileqq/files/ae/camera/capture",
				"/sdcard/Android/data/com.tencent.mobileqq/files/tbslog/txt",
				"/sdcard/Android/data/com.tencent.mobileqq/files/tencent/tbs_live_log",
				"/sdcard/Android/data/com.tencent.mobileqq/files/tencent/tbs_common_log",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/QQfile_recv/.tmp/edit_video",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/TMAssistantSDK/Download/com.tencent.mobileqq",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq/Tencent/mini/files",
				"/storage/emulated/%s/Tencent/blob/mqq",
				"/storage/emulated/%s/Tencent/ams/cache",
				"/storage/emulated/%s/Tencent/com.tencent.weread/euplog.txt",
				"/storage/emulated/%s/Tencent/imsdkvideocache",
				"/storage/emulated/%s/Tencent/Midas/Log",
				"/storage/emulated/%s/Tencent/MobileQQ/bless",
				"/data/data/com.tencent.mobileqq/files/group_catalog_temp",
				"/data/data/com.tencent.mobileqq/files/hippy/codecache",
				"/data/data/com.tencent.mobileqq/files/hippy/bundle",
				"/data/data/com.tencent.mobileqq/files/mini",
			},
			Regexes: []string{`^.*$`},
		},
	}

	// 增强清理规则
	enhancedRules = []CleanRule{
		{
			Name: "QQ核心数据清理",
			Paths: []string{
				"/data/user/%s/com.tencent.mobileqq",
				"/data/user/%s/com.tencent.tim",
				"/storage/emulated/%s/Android/data/com.tencent.mobileqq",
				"/storage/emulated/%s/Android/data/com.tencent.tim",
			},
			Regexes: []string{`^.*$`},
			Exclusions: []string{
				"files", "databases", "data", // 保留目录
				"ConfigStore2.dat", "kc", "Properties", "set_sp",
				"custom_background", "mmkv", "BusinessInfoCheckUpdateItem_new_switch_.*",
				"qa_misc", "qa_mmkv", "xa_mmkv", ".fun", // 保留文件/目录
				`Tencent/(MobileQQ|Tim|QQfile_recv|TIMfile_recv)`, // 保留关键目录
				`chatpic`, `photo`, `shortvideo`, // 保留媒体目录
				`^\d+$`, // 保留数字目录
			},
		},
	}
)

func main() {
	loadConfig()
	initLog()
	defer closeLog()

	loadWhitelist()
	loadTimeBarrier()
	if timeBarrierDays > 0 {
		writeLog(fmt.Sprintf("[时序屏障] 启用 %d 天保护，只清理 %d 天前的文件\n", timeBarrierDays, timeBarrierDays))
	} else {
		writeLog("[时序屏障] 未启用，将清理所有符合条件的文件\n")
	}

	start := time.Now()
	var totalSize int64

	// force无视 enabled 开关强制普通清理
	cleanOn := cfg.AppClean.Qq.Enabled || (len(os.Args) > 1 && os.Args[1] == "force")
	enhancedOn := cfg.AppClean.Qq.Enhanced

	// 普通清理
	if cleanOn {
		writeLog("开始普通清理")
		totalSize += processClean(normalRules, allUsers())
	}

	// 增强清理
	if enhancedOn {
		writeLog("开始增强清理")
		totalSize += processClean(enhancedRules, allUsers())
	}

	elapsed := time.Since(start)
	if cleanOn || enhancedOn {
		updateStatistics(totalSize)
		writeLog(fmt.Sprintf("[总结] 共清理 %.2f MB，耗时 %v\n\n",
			float64(totalSize)/1024/1024, elapsed.Round(time.Second)))
	} else {
		writeLog(fmt.Sprintf("[总结] 清理功能已关闭，耗时 %v\n\n", elapsed.Round(time.Second)))
	}
}

// 用户
func allUsers() []string {
	set := map[string]bool{"0": true} // 主用户兜底
	for _, pattern := range []string{"/data/user/*", "/storage/emulated/*"} {
		dirs, _ := filepath.Glob(pattern)
		for _, d := range dirs {
			u := filepath.Base(d)
			if _, err := strconv.Atoi(u); err == nil { // 仅数字目录才是 user id
				set[u] = true
			}
		}
	}
	users := make([]string, 0, len(set))
	for u := range set {
		users = append(users, u)
	}
	return users
}

func processClean(rules []CleanRule, users []string) int64 {
	var totalSize int64
	for _, rule := range rules {
		// 正则只需按规则编译一次，避免每个用户重复编译
		regexes := compileRegexes(rule.Regexes)
		exclusions := compileRegexes(rule.Exclusions)

		for _, user := range users {
			writeLog(fmt.Sprintf("[规则] %s (用户%s)\n", rule.Name, user))

			for _, path := range generatePaths(rule.Paths, user) {
				if isWhitelisted(path) {
					writeLog(fmt.Sprintf("[跳过] 路径在白名单中: %s\n", path))
					continue
				}
				if !pathExists(path) {
					continue
				}

				filepath.WalkDir(path, func(subPath string, d fs.DirEntry, err error) error {
					// 文件不能返回 SkipDir，否则同目录剩余条目会被连带跳过
					if err != nil {
						if d != nil && d.IsDir() {
							return fs.SkipDir
						}
						return nil
					}
					if isMatch(subPath, exclusions) {
						if d.IsDir() {
							return fs.SkipDir // 排除整棵子树
						}
						return nil
					}
					if isWhitelisted(subPath) {
						if d.IsDir() {
							return fs.SkipDir
						}
						return nil
					}
					if isMatch(subPath, regexes) && !isProtected(subPath) {
						if !isOlderThanBarrier(subPath, d) {
							return nil // 跳过受时序屏障保护的文件
						}
						size, err := deleteEntry(subPath, d)
						if err == nil {
							totalSize += size
							logAction(rule.Name, subPath)
							if d.IsDir() {
								return fs.SkipDir // 目录已整删，不再下降
							}
						}
					}
					return nil
				})
			}
		}
	}
	return totalSize
}

func pathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		writeLog(fmt.Sprintf("[无此路径] %s\n", path))
		return false
	}
	return true
}

func isMatch(path string, regexes []*regexp.Regexp) bool {
	normalized := filepath.ToSlash(path)
	for _, re := range regexes {
		if re.MatchString(normalized) {
			return true
		}
	}
	return false
}

// 数据保护
func isProtected(path string) bool {
	switch {
	case strings.HasSuffix(path, ".db"):
		return true
	case strings.Contains(path, "/databases"):
		return true
	}
	return false
}

func deleteEntry(path string, d fs.DirEntry) (int64, error) {
	if d.IsDir() {
		size, err := dirSize(path)
		if err != nil {
			writeLog(fmt.Sprintf("[目录错误] %s: %v\n", path, err))
			return 0, err
		}
		if err := os.RemoveAll(path); err != nil {
			writeLog(fmt.Sprintf("[删除失败] 目录 %s: %v\n", path, err))
			return 0, err
		}
		return size, nil
	}

	info, err := d.Info()
	if err != nil {
		writeLog(fmt.Sprintf("[文件错误] %s: %v\n", path, err))
		return 0, err
	}
	if err := os.Remove(path); err != nil {
		writeLog(fmt.Sprintf("[删除失败] 文件 %s: %v\n", path, err))
		return 0, err
	}
	return info.Size(), nil
}

func compileRegexes(patterns []string) []*regexp.Regexp {
	regexes := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			writeLog(fmt.Sprintf("[正则错误] %s: %v\n", pattern, err))
			continue
		}
		regexes = append(regexes, re)
	}
	return regexes
}

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			if info, err := d.Info(); err == nil {
				size += info.Size()
			}
		}
		return nil
	})
	return size, err
}

func generatePaths(templates []string, user string) []string {
	seen := make(map[string]bool)
	paths := make([]string, 0, len(templates))
	for _, tpl := range templates {
		path := tpl
		if strings.Contains(tpl, "%s") {
			path = fmt.Sprintf(tpl, user)
		}
		if !seen[path] {
			paths = append(paths, path)
			seen[path] = true
		}
	}
	return paths
}

// 更新清理统计
// basis.prop 写入互斥，防止并发覆盖统计
func lockBasis() *os.File {
	f, err := os.OpenFile(basisLockFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil // 退化为无锁
	}
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
	return f
}

func unlockBasis(f *os.File) {
	if f == nil {
		return
	}
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	_ = f.Close()
}

func updateStatistics(addedBytes int64) {
	lock := lockBasis()
	defer unlockBasis(lock)

	stats := map[string]string{}
	if data, err := os.ReadFile(statisticsFile); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
				stats[parts[0]] = parts[1]
			}
		}
	}

	addedMB := int(addedBytes / 1024 / 1024)
	stats[statsKey] = strconv.Itoa(atoi(stats[statsKey]) + 1)
	stats["statistics"] = strconv.Itoa(atoi(stats["statistics"]) + addedMB)
	stats["statistics_today"] = strconv.Itoa(atoi(stats["statistics_today"]) + addedMB)

	var content strings.Builder
	for k, v := range stats {
		content.WriteString(k + "=" + v + "\n")
	}
	if err := os.WriteFile(statisticsFile, []byte(content.String()), 0644); err != nil {
		writeLog(fmt.Sprintf("[统计错误] 写入失败: %v\n", err))
	}
}

// 日志
var (
	logEnabled bool
	logHandle  *os.File
	logBuf     *bufio.Writer
)

func initLog() {
	logEnabled = cfg.General.Log
	if !logEnabled {
		return
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		logEnabled = false
		return
	}
	logFile := filepath.Join(logDir, time.Now().Local().Format("2006-01-02")+".log")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logEnabled = false
		return
	}
	logHandle = f
	logBuf = bufio.NewWriter(f)
}

func writeLog(msg string) {
	if !logEnabled {
		return
	}
	ts := time.Now().Local().Format("15:04:05")
	logBuf.WriteString("[" + ts + "] " + logTag + msg)
}

func logAction(ruleName, path string) {
	writeLog(fmt.Sprintf("[清理] %s：%s\n", ruleName, path))
}

func closeLog() {
	if logBuf != nil {
		logBuf.Flush()
	}
	if logHandle != nil {
		logHandle.Close()
	}
}

func atoi(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}
