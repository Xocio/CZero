package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

var errAlreadyGone = errors.New("source already gone")

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
	appRulesCfg  = moduleBase + "/list/app.json"

	recycleDir      = "/storage/emulated/0/Recycle"
	recycleDataDir  = "files"
	manifestName    = "manifest.json"
	recycleKeepDays = 7
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

// app.json App
type appRule struct {
	Package string   `json:"package"`
	Name    string   `json:"name"`
	Enabled bool     `json:"enabled"`
	Paths   []string `json:"paths"`
}

type appRulesFile struct {
	Version int       `json:"version"`
	Apps    []appRule `json:"apps"`
}

func loadAppRules(log *logger) []string {
	data, err := os.ReadFile(appRulesCfg)
	if err != nil {
		return nil
	}
	var f appRulesFile
	if err := json.Unmarshal(data, &f); err != nil {
		log.log("app.json 解析失败: %v", err)
		return nil
	}
	var out []string
	for _, a := range f.Apps {
		if !a.Enabled {
			log.log("app.json 跳过已停用应用: %s", a.Package)
			continue
		}
		for _, p := range a.Paths {
			// 只接受绝对路径，防御手改出相对路径误清工作目录
			if strings.HasPrefix(p, "/") {
				out = append(out, p)
			}
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
				log.log("权限不足跳过: %s", path)
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

// 回收站
type recycleItem struct {
	Original string `json:"original"` // 原始绝对路径
	Size     int64  `json:"size"`
}

type recycleManifest struct {
	ID         string        `json:"id"`
	CreatedAt  int64         `json:"created_at"` // Unix 秒
	ExpireAt   int64         `json:"expire_at"`  // Unix 秒，超过即可清除
	TotalBytes int64         `json:"total_bytes"`
	Items      []recycleItem `json:"items"`
}

type recycleBin struct {
	mu       sync.Mutex
	inited   bool
	initErr  error
	session  string // recycleDir/<id>
	filesDir string // session/files
	m        recycleManifest
	log      *logger
}

func newRecycleBin(log *logger) *recycleBin {
	return &recycleBin{log: log}
}

// 会话目录延迟创建，本次运行没有清出任何文件时不留空目录
func (b *recycleBin) ensureInit() error {
	if b.inited {
		return b.initErr
	}
	b.inited = true

	id := time.Now().Local().Format("20060102-150405") + "-" + strconv.Itoa(os.Getpid())
	session := filepath.Join(recycleDir, id)
	filesDir := filepath.Join(session, recycleDataDir)
	if err := os.MkdirAll(filesDir, 0755); err != nil {
		b.initErr = err
		return err
	}
	// 阻止媒体库扫描回收站内容
	_ = os.WriteFile(filepath.Join(recycleDir, ".nomedia"), nil, 0644)

	now := time.Now()
	b.session = session
	b.filesDir = filesDir
	b.m = recycleManifest{
		ID:        id,
		CreatedAt: now.Unix(),
		ExpireAt:  now.AddDate(0, 0, recycleKeepDays).Unix(),
	}
	return nil
}

// moveToRecycle 把文件移动到回收站内的镜像路径
func (b *recycleBin) moveToRecycle(path string, size int64) error {
	b.mu.Lock()
	if err := b.ensureInit(); err != nil {
		b.mu.Unlock()
		return err
	}
	filesDir := b.filesDir
	b.mu.Unlock()

	dst := filepath.Join(filesDir, strings.TrimPrefix(path, "/"))
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	if err := movePath(path, dst); err != nil {
		return err
	}

	b.mu.Lock()
	b.m.Items = append(b.m.Items, recycleItem{Original: path, Size: size})
	b.m.TotalBytes += size
	b.mu.Unlock()
	return nil
}

func (b *recycleBin) saveManifest() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.session == "" {
		return
	}
	if len(b.m.Items) == 0 {
		// 没有实际移动任何文件，回收会话目录
		_ = os.RemoveAll(b.session)
		return
	}
	data, err := json.MarshalIndent(b.m, "", "  ")
	if err != nil {
		return
	}
	if err := os.WriteFile(filepath.Join(b.session, manifestName), data, 0644); err != nil {
		b.log.log("写入回收站清单失败: %s → %v", b.session, err)
	}
}

// movePath 移动单个文件，优先 rename，跨文件系统回退为复制后删源
func movePath(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	} else if os.IsNotExist(err) {
		return errAlreadyGone
	}

	fi, err := os.Lstat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return errAlreadyGone
		}
		return err
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		if err != nil {
			if os.IsNotExist(err) {
				return errAlreadyGone
			}
			return err
		}
		_ = os.Remove(dst)
		if err := os.Symlink(target, dst); err != nil {
			return err
		}
		if err := os.Remove(src); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	if fi.IsDir() {
		return fmt.Errorf("跨文件系统移动目录不受支持: %s", src)
	}

	in, err := os.Open(src)
	if err != nil {
		if os.IsNotExist(err) {
			return errAlreadyGone
		}
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fi.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		_ = os.Remove(dst)
		return err
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(dst)
		return err
	}
	// 保留修改时间，避免恢复后时序屏障判断失真
	_ = os.Chtimes(dst, fi.ModTime(), fi.ModTime())
	// 拷贝已完成，源文件此时才消失不影响结果
	if err := os.Remove(src); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// purgeExpired 清除超过保留期的回收会话
func purgeExpired(log *logger) {
	entries, err := os.ReadDir(recycleDir)
	if err != nil {
		return
	}
	now := time.Now().Unix()
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		session := filepath.Join(recycleDir, e.Name())
		expireAt := int64(0)
		if m, err := loadManifest(session); err == nil {
			expireAt = m.ExpireAt
		} else if fi, err := e.Info(); err == nil {
			// 清单缺失时按目录时间兜底
			expireAt = fi.ModTime().AddDate(0, 0, recycleKeepDays).Unix()
		} else {
			continue
		}
		if now < expireAt {
			continue
		}
		if err := os.RemoveAll(session); err != nil {
			log.log("清除过期回收会话失败: %s → %v", session, err)
		} else {
			log.log("已清除过期回收会话: %s", e.Name())
		}
	}
}

func loadManifest(session string) (recycleManifest, error) {
	var m recycleManifest
	data, err := os.ReadFile(filepath.Join(session, manifestName))
	if err != nil {
		return m, err
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return m, err
	}
	return m, nil
}

// restoreSession 把一个回收会话内的所有文件原样移回原始位置
func restoreSession(session string, log *logger) (restored, failed int) {
	filesDir := filepath.Join(session, recycleDataDir)
	_ = filepath.WalkDir(filesDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(filesDir, path)
		if err != nil {
			failed++
			return nil
		}
		orig := "/" + filepath.ToSlash(rel)
		if err := os.MkdirAll(filepath.Dir(orig), 0755); err != nil {
			log.log("恢复失败(建目录): %s → %v", orig, err)
			failed++
			return nil
		}
		if err := movePath(path, orig); err != nil {
			log.log("恢复失败: %s → %v", orig, err)
			failed++
			return nil
		}
		restored++
		return nil
	})
	if failed == 0 {
		_ = os.RemoveAll(session)
	}
	return restored, failed
}

// runRestore 处理 restore 子命令，target 为会话 ID 或 "all"，结果以 JSON 输出到 stdout 供 app 读取
func runRestore(target string, log *logger) {
	var sessions []string
	if target == "all" {
		if entries, err := os.ReadDir(recycleDir); err == nil {
			for _, e := range entries {
				if e.IsDir() {
					sessions = append(sessions, filepath.Join(recycleDir, e.Name()))
				}
			}
		}
	} else {
		session := filepath.Join(recycleDir, filepath.Base(target))
		if fi, err := os.Stat(session); err != nil || !fi.IsDir() {
			fmt.Printf(`{"ok":false,"error":"session not found: %s"}`+"\n", target)
			return
		}
		sessions = append(sessions, session)
	}

	totalRestored, totalFailed := 0, 0
	for _, s := range sessions {
		r, f := restoreSession(s, log)
		totalRestored += r
		totalFailed += f
		log.log("恢复回收会话 %s: 成功 %d, 失败 %d", filepath.Base(s), r, f)
	}
	fmt.Printf(`{"ok":%t,"restored":%d,"failed":%d}`+"\n", totalFailed == 0, totalRestored, totalFailed)
}

// sessionSummary 读取会话清单；清单缺失（清理进程被中断等）时扫描 files 目录现场合成，并写回磁盘补全，保证后续 list/purge 行为一致。会话内无文件时返回 false
func sessionSummary(session string) (recycleManifest, bool) {
	if m, err := loadManifest(session); err == nil {
		return m, true
	}

	fi, err := os.Stat(session)
	if err != nil {
		return recycleManifest{}, false
	}
	m := recycleManifest{
		ID:        filepath.Base(session),
		CreatedAt: fi.ModTime().Unix(),
		ExpireAt:  fi.ModTime().AddDate(0, 0, recycleKeepDays).Unix(),
	}

	filesDir := filepath.Join(session, recycleDataDir)
	_ = filepath.WalkDir(filesDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, e := d.Info()
		if e != nil {
			return nil
		}
		rel, e := filepath.Rel(filesDir, path)
		if e != nil {
			return nil
		}
		m.Items = append(m.Items, recycleItem{Original: "/" + filepath.ToSlash(rel), Size: info.Size()})
		m.TotalBytes += info.Size()
		return nil
	})
	if len(m.Items) == 0 {
		return recycleManifest{}, false
	}

	if data, err := json.MarshalIndent(m, "", "  "); err == nil {
		_ = os.WriteFile(filepath.Join(session, manifestName), data, 0644)
	}
	return m, true
}

// runList 输出所有回收会话的清单 JSON 数组，供 app 展示
func runList() {
	list := []recycleManifest{}
	if entries, err := os.ReadDir(recycleDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			if m, ok := sessionSummary(filepath.Join(recycleDir, e.Name())); ok {
				list = append(list, m)
			}
		}
	}
	data, _ := json.Marshal(list)
	fmt.Println(string(data))
}

// 清理
type cleaner struct {
	barrierDays int
	cutoff      time.Time
	wl          *whitelist
	bin         *recycleBin
	log         *logger
	totalBytes  atomic.Int64
}

func newCleaner(barrierDays int, wl *whitelist, bin *recycleBin, log *logger) *cleaner {
	c := &cleaner{
		barrierDays: barrierDays,
		wl:          wl,
		bin:         bin,
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
	if err := c.bin.moveToRecycle(path, size); err != nil {
		if !errors.Is(err, errAlreadyGone) {
			c.log.log("移入回收站失败: %s → %v", path, err)
		}
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
		// 回收站自身永不清理，防止规则匹配到 /storage/emulated/0 时套娃
		if path == recycleDir {
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

		if err := c.bin.moveToRecycle(path, st.Size); err == nil {
			cleaned += st.Size
		} else if !errors.Is(err, errAlreadyGone) {
			c.log.log("移入回收站失败: %s → %v", path, err)
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
		if p == recycleDir || strings.HasPrefix(p, recycleDir+"/") {
			c.log.log("跳过回收站自身: %s", p)
			continue
		}
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

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "restore":
			if len(os.Args) < 3 {
				fmt.Println(`{"ok":false,"error":"usage: restore <session-id|all>"}`)
				return
			}
			runRestore(os.Args[2], log)
			log.flush()
			return
		case "list":
			runList()
			return
		case "purge":
			purgeExpired(log)
			log.flush()
			return
		}
	}

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
	log.log("共加载清理规则: %d 条", len(patterns))

	// app.json
	if appPatterns := loadAppRules(log); len(appPatterns) > 0 {
		log.log("共加载应用缓存规则(app.json): %d 条", len(appPatterns))
		patterns = append(patterns, appPatterns...)
	}
	if len(patterns) == 0 {
		log.log("清理路径文件为空或不存在")
		return
	}

	// 清除超过保留期的历史回收会话
	purgeExpired(log)

	bin := newRecycleBin(log)
	c := newCleaner(barrierDays, wl, bin, log)
	totalBytes := c.run(patterns)
	bin.saveManifest()
	if n := len(bin.m.Items); n > 0 {
		log.log("已移入回收站: %d 个文件 → %s (保留 %d 天)", n, bin.session, recycleKeepDays)
	}

	mb := totalBytes / (1024 * 1024)
	if mb > 0 {
		if err := updateStats(int(mb), log); err != nil {
			log.log("统计更新失败: %v", err)
		}
	}

	log.log("本次清理总大小: %d MB", mb)
	log.log("其他缓存清理完成")
}
