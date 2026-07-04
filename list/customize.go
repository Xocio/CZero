package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// 路径常量
const (
	moduleBase   = "/data/adb/modules/CZero"
	logDir       = moduleBase + "/log"
	logTag       = "[自定义] "
	cfgPath      = moduleBase + "/config.json"
	pathsCfg     = moduleBase + "/list/clean_paths.prop"
	basisCfg     = moduleBase + "/basis/basis.prop"
	basisLockCfg = moduleBase + "/basis/.basis.lock"
	whitelistCfg = moduleBase + "/list/clean_whitelist.prop"
)

// 日志
type logger struct {
	mu      sync.Mutex
	file    *os.File
	writer  *bufio.Writer
	enabled bool
}

func newLogger(enabled bool) *logger {
	l := &logger{enabled: enabled}
	if !enabled {
		return l
	}
	_ = os.MkdirAll(logDir, 0755)
	logFile := logDir + "/" + time.Now().Local().Format("2006-01-02") + ".log" 
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		l.enabled = false
		return l
	}
	l.file = f
	l.writer = bufio.NewWriterSize(f, 8192)
	return l
}

func (l *logger) close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.writer != nil {
		_ = l.writer.Flush()
	}
	if l.file != nil {
		_ = l.file.Close()
	}
}

func (l *logger) log(format string, a ...any) {
	if !l.enabled || l.writer == nil {
		return
	}
	ts := time.Now().Local().Format("15:04:05")
	msg := fmt.Sprintf(format, a...)
	line := "[" + ts + "] " + logTag + msg + "\n"

	l.mu.Lock()
	_, _ = l.writer.WriteString(line)
	l.mu.Unlock()
}

func (l *logger) flush() {
	if !l.enabled || l.writer == nil {
		return
	}
	l.mu.Lock()
	_ = l.writer.Flush()
	l.mu.Unlock()
}

// 配置读取
func readCfgMap(path string) map[string]string {
	m := make(map[string]string, 32)
	f, err := os.Open(path)
	if err != nil {
		return m
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		if i := strings.IndexByte(line, '='); i > 0 {
			m[strings.TrimSpace(line[:i])] = strings.TrimSpace(line[i+1:])
		}
	}
	return m
}

// JSON 配置
type Config struct {
	General struct {
		Log                 bool `json:"log"`
		TemporalBarrierDays int  `json:"temporal_barrier_days"`
	} `json:"general"`
	AppClean struct {
		Other struct {
			Enabled bool `json:"enabled"`
		} `json:"other"`
	} `json:"app_clean"`
}

func loadConfig(path string) Config {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}
	_ = json.Unmarshal(data, &cfg)
	return cfg
}

func loadLines(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var out []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 4096), 512*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line != "" && line[0] != '#' {
			out = append(out, line)
		}
	}
	return out
}

// Glob 展开
func hasGlobMeta(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

func globExpand(pattern string, log *logger) []string {
	pattern = filepath.Clean(pattern)

	if !hasGlobMeta(pattern) {
		if _, err := os.Lstat(pattern); err == nil {
			return []string{pattern}
		}
		return nil
	}

	if strings.Contains(pattern, "**") {
		return globDoublestar(pattern, log)
	}

	if matches, err := filepath.Glob(pattern); err == nil && len(matches) > 0 {
		return matches
	}

	return globSegments(pattern, log)
}

func globDoublestar(pattern string, log *logger) []string {
	idx := strings.Index(pattern, "**")
	base := filepath.Clean(pattern[:idx])
	suffix := strings.TrimPrefix(pattern[idx+2:], "/")
	if base == "" || base == "." {
		base = "/"
	}

	var results []string
	_ = filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				log.log("⚠️ 权限不足跳过: %s", path)
			}
			return filepath.SkipDir
		}
		if !d.IsDir() {
			return nil
		}
		if suffix == "" {
			results = append(results, path)
			return nil
		}
		candidate := filepath.Join(path, suffix)
		if hasGlobMeta(suffix) {
			if sub, e := filepath.Glob(candidate); e == nil {
				results = append(results, sub...)
			}
		} else {
			if _, e := os.Lstat(candidate); e == nil {
				results = append(results, candidate)
			}
		}
		return nil
	})
	return results
}

func globSegments(pattern string, log *logger) []string {
	clean := filepath.Clean(pattern)
	isAbs := filepath.IsAbs(clean)
	segments := splitPath(clean)

	var candidates []string
	if isAbs {
		candidates = []string{"/"}
	} else {
		candidates = []string{"."}
	}

	for _, seg := range segments {
		if len(candidates) == 0 {
			break
		}

		var next []string

		if !hasGlobMeta(seg) {
			for _, c := range candidates {
				p := filepath.Join(c, seg)
				if _, err := os.Lstat(p); err == nil {
					next = append(next, p)
				}
			}
		} else {
			for _, c := range candidates {
				entries, err := os.ReadDir(c)
				if err != nil {
					if os.IsPermission(err) {
						log.log("权限不足跳过目录: %s", c)
					}
					continue
				}
				for _, e := range entries {
					name := e.Name()
					if matched, _ := filepath.Match(seg, name); matched {
						next = append(next, filepath.Join(c, name))
					}
				}
			}
		}

		candidates = next
	}

	return candidates
}

func splitPath(p string) []string {
	raw := strings.Split(filepath.Clean(p), string(os.PathSeparator))
	var out []string
	for _, s := range raw {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// 去重与嵌套过滤
func dedup(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	// 先 Clean 再排序，保证后续的父子前缀去重不变量成立
	for i := range paths {
		paths[i] = filepath.Clean(paths[i])
	}
	sort.Strings(paths)

	var out []string
	prev := ""
	for _, p := range paths {
		if prev != "" && strings.HasPrefix(p, prev+"/") {
			continue
		}
		out = append(out, p)
		prev = p
	}
	return out
}

// 白名单
type whitelist struct {
	rules []string
}

func newWhitelist(lines []string) *whitelist {
	return &whitelist{rules: lines}
}

func (w *whitelist) blocked(path string) bool {
	for _, rule := range w.rules {
		if hasGlobMeta(rule) {
			if ok, _ := filepath.Match(rule, path); ok {
				return true
			}
			if ok, _ := filepath.Match(rule, filepath.Base(path)); ok {
				return true
			}
			continue
		}
		if strings.Contains(path, rule) { // 子串匹配已涵盖前缀
			return true
		}
	}
	return false
}

// 清理
type cleaner struct {
	barrierDays int
	cutoff      time.Time
	wl          *whitelist
	log         *logger
	totalBytes  atomic.Int64
}

func newCleaner(barrierDays int, wl *whitelist, log *logger) *cleaner {
	c := &cleaner{
		barrierDays: barrierDays,
		wl:          wl,
		log:         log,
	}
	if barrierDays > 0 {
		c.cutoff = time.Now().AddDate(0, 0, -barrierDays)
	}
	return c
}

func (c *cleaner) shouldDelete(modTime time.Time) bool {
	if c.barrierDays <= 0 {
		return true
	}
	return modTime.Before(c.cutoff)
}

func (c *cleaner) cleanSingleFile(path string) int64 {
	var st syscall.Stat_t
	if err := syscall.Lstat(path, &st); err != nil {
		c.log.log("无法访问文件: %s → %v", path, err)
		return 0
	}
	mtime := time.Unix(st.Mtim.Sec, st.Mtim.Nsec)
	if !c.shouldDelete(mtime) {
		return 0
	}
	size := st.Size
	if err := os.Remove(path); err != nil {
		c.log.log("删除文件失败: %s → %v", path, err)
		return 0
	}
	return size
}

// cleanDir 跳过白名单子树，按时序屏障逐文件删除并累加大小，最后清掉空目录
// 单趟遍历边删边累加，避免先 dirSize 再 RemoveAll 的统计竞争
func (c *cleaner) cleanDir(dirPath string) int64 {
	var cleaned int64
	_ = filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				c.log.log("权限不足跳过: %s", path)
			}
			return filepath.SkipDir
		}
		// 白名单保护子路径
		if c.wl.blocked(path) {
			c.log.log("白名单跳过: %s", path)
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}

		var st syscall.Stat_t
		if err := syscall.Lstat(path, &st); err != nil {
			return nil
		}
		mtime := time.Unix(st.Mtim.Sec, st.Mtim.Nsec)
		if !c.shouldDelete(mtime) {
			return nil
		}

		if err := os.Remove(path); err == nil {
			cleaned += st.Size
		} else {
			c.log.log("删除失败: %s → %v", path, err)
		}
		return nil
	})

	removeEmptyDirs(dirPath)
	// 顶层目录若已清空则一并删除
	if entries, err := os.ReadDir(dirPath); err == nil && len(entries) == 0 {
		_ = os.Remove(dirPath)
	}
	return cleaned
}

func (c *cleaner) cleanPath(path string) int64 {
	fi, err := os.Lstat(path)
	if err != nil {
		c.log.log("无法访问: %s → %v", path, err)
		return 0
	}
	if !fi.IsDir() {
		sz := c.cleanSingleFile(path)
		if sz > 0 {
			c.log.log("清理文件: %s (%d KB)", path, sz/1024)
		}
		return sz
	}
	sz := c.cleanDir(path)
	if sz > 0 {
		c.log.log("清理目录: %s (%d MB)", path, sz/1024/1024)
	}
	return sz
}

func (c *cleaner) run(patterns []string) int64 {
	var allPaths []string
	for _, pat := range patterns {
		expanded := globExpand(pat, c.log)
		if len(expanded) == 0 {
			c.log.log("无匹配: %s", pat)
			continue
		}
		allPaths = append(allPaths, expanded...)
	}

	allPaths = dedup(allPaths)

	// 顶层白名单过滤
	var targets []string
	for _, p := range allPaths {
		if c.wl.blocked(p) {
			c.log.log("白名单跳过: %s", p)
			continue
		}
		targets = append(targets, p)
	}

	if len(targets) == 0 {
		c.log.log("无需清理的路径")
		return 0
	}

	// 并发清理，限流
	procs := runtime.NumCPU()
	if procs > 4 {
		procs = 4
	}
	sem := make(chan struct{}, procs)
	var wg sync.WaitGroup

	for _, p := range targets {
		wg.Add(1)
		sem <- struct{}{}
		go func(path string) {
			defer wg.Done()
			defer func() { <-sem }()
			if sz := c.cleanPath(path); sz > 0 {
				c.totalBytes.Add(sz)
			}
		}(p)
	}

	wg.Wait()
	c.log.flush()
	return c.totalBytes.Load()
}

// 目录工具
func removeEmptyDirs(root string) {
	var dirs []string
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && path != root {
			dirs = append(dirs, path)
		}
		return nil
	})

	sort.Slice(dirs, func(i, j int) bool {
		return len(dirs[i]) > len(dirs[j])
	})

	for _, d := range dirs {
		entries, err := os.ReadDir(d)
		if err == nil && len(entries) == 0 {
			_ = os.Remove(d)
		}
	}
}

// 统计
func updateStats(addMB int, log *logger) error {
	if addMB <= 0 {
		return nil
	}

	// basis.prop 写入互斥，防止并发覆盖统计
	if lf, err := os.OpenFile(basisLockCfg, os.O_CREATE|os.O_RDWR, 0644); err == nil {
		_ = syscall.Flock(int(lf.Fd()), syscall.LOCK_EX)
		defer func() { _ = syscall.Flock(int(lf.Fd()), syscall.LOCK_UN); _ = lf.Close() }()
	}

	kv := readCfgMap(basisCfg)
	old, _ := strconv.Atoi(kv["statistics"])
	today, _ := strconv.Atoi(kv["statistics_today"])
	kv["statistics"] = strconv.Itoa(old + addMB)
	kv["statistics_today"] = strconv.Itoa(today + addMB)

	_ = os.MkdirAll(filepath.Dir(basisCfg), 0755)

	tmp := basisCfg + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}

	// 排序一下key
	keys := make([]string, 0, len(kv))
	for k := range kv {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	w := bufio.NewWriter(f)
	for _, k := range keys {
		_, _ = w.WriteString(k + "=" + kv[k] + "\n")
	}
	_ = w.Flush()
	_ = f.Close()

	if err := os.Rename(tmp, basisCfg); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename 失败: %w", err)
	}

	log.log("统计已更新: +%d MB (累计: %s MB)", addMB, kv["statistics"])
	return nil
}

// main
func main() {
	// 自适应核心数，最高4，其实影响很小
	procs := runtime.NumCPU()
	if procs > 4 {
		procs = 4
	}
	runtime.GOMAXPROCS(procs)

	cfg := loadConfig(cfgPath)
	log := newLogger(cfg.General.Log)
	defer log.close()

	log.log("开始其他缓存清理 (GOMAXPROCS=%d)", procs)

	barrierDays := cfg.General.TemporalBarrierDays
	if barrierDays > 0 {
		log.log("时序屏障已启用：只清理 %d 天前的文件", barrierDays)
	} else {
		log.log("时序屏障未启用：将清理所有匹配的文件")
	}

	// force无视 enabled 开关
	force := len(os.Args) > 1 && os.Args[1] == "force"
	if !cfg.AppClean.Other.Enabled && !force {
		log.log("清理功能未启用，退出程序")
		return
	}

	wlLines := loadLines(whitelistCfg)
	wl := newWhitelist(wlLines)
	log.log("共加载白名单路径: %d 个", len(wlLines))

	patterns := loadLines(pathsCfg)
	if len(patterns) == 0 {
		log.log("清理路径文件为空或不存在")
		return
	}
	log.log("共加载清理规则: %d 条", len(patterns))

	c := newCleaner(barrierDays, wl, log)
	totalBytes := c.run(patterns)

	mb := totalBytes / (1024 * 1024)
	if mb > 0 {
		if err := updateStats(int(mb), log); err != nil {
			log.log("统计更新失败: %v", err)
		}
	}

	log.log("本次清理总大小: %d MB", mb)
	log.log("其他缓存清理完成")
}
