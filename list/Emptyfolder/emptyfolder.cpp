#include <iostream>
#include <fstream>
#include <string>
#include <vector>
#include <filesystem>
#include <ctime>
#include <algorithm>
#include <unordered_map>
#include <iterator>
#include <cstdlib>
#include <cstdio>
#include "common/json_config.h"
#include "common/file_lock.h"

namespace fs = std::filesystem;

//通知
static void notifyIsland(const char* phase, const char* action, int rc) {
    const char* no = getenv("CZERO_NO_NOTIFY");
    if (no && no[0] == '1') return;
    const char* recv = "com.web.czero/com.web.czero.core.notify.CleanEventReceiver";
    char cmd[320];
    if (std::string(phase) == "done")
        snprintf(cmd, sizeof(cmd),
                 "am broadcast -n %s --es phase done --es action %s --ei ok %d >/dev/null 2>&1",
                 recv, action, rc);
    else
        snprintf(cmd, sizeof(cmd),
                 "am broadcast -n %s --es phase start --es action %s >/dev/null 2>&1",
                 recv, action);
    int r = system(cmd); (void)r;
}

const std::string MODULE_PATH = "/data/adb/modules/CZero/";
const std::string CONFIG_FILE  = MODULE_PATH + "config.json";
const std::string LOG_DIR      = MODULE_PATH + "log/";
const std::string LOG_TAG      = "[空文件夹] ";
const std::string BASIS_FILE   = MODULE_PATH + "basis/basis.prop";
const std::string DIR_LIST     = MODULE_PATH + "list/Emptyfolder/directories.prop";
const std::string WHITELIST    = MODULE_PATH + "list/Emptyfolder/emptyfolder_white.prop";

static std::vector<std::string> logBuffer;
static bool loggingEnabled = false;

// JSON 解析
using czero::parseJsonConfig;

static std::vector<std::string> readLines(const std::string& path) {
    std::vector<std::string> lines;
    std::ifstream file(path); if (!file.is_open()) return lines;
    std::string line;
    while (std::getline(file, line)) {
        size_t start = line.find_first_not_of(" \t\r\n");
        if (start == std::string::npos || line[start] == '#') continue;
        size_t end = line.find_last_not_of(" \t\r\n");
        lines.push_back(line.substr(start, end - start + 1));
    }
    return lines;
}

static void flushLog() {
    if (logBuffer.empty()) return;
    time_t now = time(nullptr); char day[16];
    strftime(day, sizeof(day), "%Y-%m-%d", localtime(&now));   // 当天日期作文件名
    std::ofstream f(LOG_DIR + day + ".log", std::ios::app);
    if (f.is_open()) for (const auto& msg : logBuffer) f << msg << '\n';
    logBuffer.clear();
}

static void addLog(const std::string& msg) {
    if (!loggingEnabled) return;
    time_t now = time(nullptr); char buf[32];
    strftime(buf, sizeof(buf), "%H:%M:%S", localtime(&now));
    logBuffer.push_back(std::string("[") + buf + "] " + LOG_TAG + msg);
    if (logBuffer.size() >= 100) flushLog();
}

static bool matchWildcard(const std::string& pattern, const std::string& path) {
    size_t plen = pattern.length(), slen = path.length();
    if (plen == 0) return slen == 0;
    bool sf = (pattern.front() == '*'), sb = (pattern.back() == '*');
    if (sf && sb && plen > 1) return path.find(pattern.substr(1, plen - 2)) != std::string::npos;
    if (sf) { std::string s = pattern.substr(1); return slen >= s.length() && path.compare(slen - s.length(), s.length(), s) == 0; }
    if (sb) { std::string p = pattern.substr(0, plen - 1); return path.compare(0, p.length(), p) == 0; }
    return pattern == path;
}

class WhitelistChecker {
    std::vector<std::string> exactOrPrefix, wildcards;
    static std::string normalize(const std::string& p) { std::string s = p; while (s.length() > 1 && s.back() == '/') s.pop_back(); return s; }
    static bool isPrefixOf(const std::string& base, const std::string& path) {
        if (path == base) return true;
        return path.length() > base.length() && path.compare(0, base.length(), base) == 0 && path[base.length()] == '/';
    }
public:
    void init(const std::vector<std::string>& whitelist) {
        std::vector<std::string> all = {"/storage/emulated/0/Android/data/com.ss.android.ugc.aweme"};
        all.insert(all.end(), whitelist.begin(), whitelist.end());
        for (const auto& e : all) { std::string n = normalize(e); if (n.empty()) continue; (n.find('*') != std::string::npos ? wildcards : exactOrPrefix).push_back(n); }
    }
    bool isInWhitelist(const std::string& path) const {
        std::string n = normalize(path);
        for (const auto& base : exactOrPrefix) if (isPrefixOf(base, n)) return true;
        for (const auto& pat : wildcards) if (matchWildcard(pat, n)) return true;
        return false;
    }
};

static void updateBasisCount(int count) {
    if (count == 0) return;
    czero::FileLock lock(czero::kBasisLock);  // 统计写入互斥
    std::ifstream in(BASIS_FILE); std::string content;
    if (in.is_open()) { content.assign(std::istreambuf_iterator<char>(in), std::istreambuf_iterator<char>()); in.close(); }
    size_t pos = content.find("emptyfolder=");
    if (pos != std::string::npos) {
        size_t ep = pos + 12, endPos = content.find('\n', ep); if (endPos == std::string::npos) endPos = content.length();
        int cur = 0; try { cur = std::stoi(content.substr(ep, endPos - ep)); } catch (...) {}
        content.replace(ep, endPos - ep, std::to_string(cur + count));
    } else { if (!content.empty() && content.back() != '\n') content += '\n'; content += "emptyfolder=" + std::to_string(count) + '\n'; }
    std::ofstream out(BASIS_FILE); if (out.is_open()) out << content;
}

static int cleanEmptyDirs(const std::string& dirPath, const WhitelistChecker& wl) {
    int cleaned = 0; std::vector<fs::path> toRemove;
    try {
        for (const auto& entry : fs::directory_iterator(dirPath, fs::directory_options::skip_permission_denied)) {
            if (!fs::is_directory(entry)) continue;
            const std::string path = entry.path().string();
            if (wl.isInWhitelist(path)) { addLog("跳过受保护目录: " + path); continue; }
            cleaned += cleanEmptyDirs(path, wl);
            try { if (fs::exists(entry) && fs::is_directory(entry) && fs::is_empty(entry)) toRemove.push_back(entry.path()); } catch (...) {}
        }
    } catch (const std::exception& e) { addLog("访问出错 " + dirPath + ": " + e.what()); }
    for (const auto& dir : toRemove) {
        try { fs::remove(dir); cleaned++; addLog("已移除: " + dir.string()); }
        catch (const std::exception& e) { addLog("移除失败 " + dir.string() + ": " + e.what()); }
    }
    return cleaned;
}

int main(int argc, char* argv[]) {
    // force无视 enabled
    bool force = (argc > 1 && std::string(argv[1]) == "force");
    auto cfg = parseJsonConfig(CONFIG_FILE);
    if (!force && cfg["empty_folder.enabled"] != "true") return 0;
    loggingEnabled = (cfg["general.log"] == "true");
    std::vector<std::string> dirs = readLines(DIR_LIST), whitelist = readLines(WHITELIST);
    WhitelistChecker wl; wl.init(whitelist);
    notifyIsland("start", "emptyfolder", 0);
    addLog("开始空目录清理"); int total = 0;
    for (const auto& dir : dirs) {
        if (!fs::exists(dir)) { addLog("目录不存在: " + dir); continue; }
        addLog("扫描: " + dir); total += cleanEmptyDirs(dir, wl);
    }
    updateBasisCount(total);
    addLog("清理完成，移除了 " + std::to_string(total) + " 个空目录");
    notifyIsland("done", "emptyfolder", 0);
    flushLog(); return 0;
}