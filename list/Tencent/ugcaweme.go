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
	logTag         = "[抖音] "
	statisticsFile = "/data/adb/modules/CZero/basis/basis.prop"
	basisLockFile  = "/data/adb/modules/CZero/basis/.basis.lock"
	whitelistFile  = "/data/adb/modules/CZero/list/clean_whitelist.prop"

	statsKey = "douyin" // 统计键名
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
		Douyin struct {
			Enabled  bool `json:"enabled"`
			Enhanced bool `json:"enhanced"`
		} `json:"douyin"`
	} `json:"app_clean"`
}

// 配置启动时一次性读入
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

// 白名单启动时一次性加载
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
			Name: "抖音基础缓存",
			Paths: []string{
				"/storage/emulated/0/Android/data/com.ss.android.ugc.aweme/cache/",
				"/data/data/com.ss.android.ugc.aweme/files/logs/",
				"/storage/emulated/0/Android/data/com.ss.android.ugc.aweme/files/perf",
				"/storage/emulated/0/ByteDownload",
				"/storage/emulated/0/Android/data/com.ss.android.ugc.aweme/bytedance",
				"/storage/emulated/0/Android/data/com.ss.android.ugc.aweme/files/MiPushLog",
				"/storage/emulated/0/Android/data/com.ss.android.ugc.aweme/files/feedback_f",
				"/storage/emulated/0/aweme_monitor",
			},
			Regexes: []string{`^.*$`},
		},
	}

	enhancedRules = []CleanRule{
		{
			Name: "抖音深度缓存清理",
			Paths: []string{
				"/data/data/com.ss.android.ugc.aweme/files/",
				"/storage/emulated/%s/Android/data/com.ss.android.ugc.aweme/",
				"/data/media/%s/Android/data/com.ss.android.ugc.aweme/",
				"/mnt/runtime/default/emulated/%s/Android/data/com.ss.android.ugc.aweme/",
				"/data/user/%s/com.ss.android.ugc.aweme/", // 多用户支持
			},
			Regexes: []string{
				`(?i)/cache/`,
				`\.tmp$`,
				`\.temp$`,
				`\.log$`,
				`_draft$`,
				`^tmp_`,
				`^draft_`,
				`^effect_tmp_`,
				`download/.+\.tmp$`,
				`template/.+\.cache$`,
				`music/.+/.+_draft$`,
				`^sticker_tmp_`,
				`^watermark_`,
				`^upload_tmp_`,
				`/data/tmp`,
				`/cache_tmp/`,
			},
			Exclusions: []string{
				`(?i)/files/main/`,
				`(?i)/beauty/`,
				`(?i)/fonts/`,
				`(?i)/user_info/`,
				`(?i)/settings/`,
				`(?i)/video_config/`,
				`aweme\.config$`,
				`shared_prefs/`,
				`databases/`,
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

	// force无视 enabled
	cleanOn := cfg.AppClean.Douyin.Enabled || (len(os.Args) > 1 && os.Args[1] == "force")
	enhancedOn := cfg.AppClean.Douyin.Enhanced

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
	writeLog(fmt.Sprintf("[清理] %s → %s\n", ruleName, path))
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
