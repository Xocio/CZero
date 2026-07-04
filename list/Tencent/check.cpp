#include <iostream>
#include <fstream>
#include <string>
#include <vector>
#include <unordered_map>
#include <unordered_set>
#include <cstdlib>
#include <cstdio>
#include <ctime>
#include <unistd.h>
#include <sys/stat.h>
#include <dirent.h>
#include <iterator>
#include "common/json_config.h"

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

const std::string LOG_DIR     = "/data/adb/modules/CZero/log/";
const std::string LOG_TAG     = "[检测] ";
const std::string CONFIG_FILE = "/data/adb/modules/CZero/config.json";
const std::string STATE_FILE  = "/data/adb/modules/CZero/basis/check.prop";

struct AppConfig { std::string package, stateKey, scriptPath, name; };

const std::vector<AppConfig> MONITORED_APPS = {
    {"com.tencent.mm",           "wechat", "/data/adb/modules/CZero/list/Tencent/tencentmm","微信"},
    {"com.tencent.mobileqq",     "qq",     "/data/adb/modules/CZero/list/Tencent/mobileqq","QQ"},
    {"com.ss.android.ugc.aweme", "douyin", "/data/adb/modules/CZero/list/Tencent/ugcaweme","抖音"}
};

const std::unordered_set<std::string> GAME_PACKAGES = {
    "com.tencent.tmgp.sgame","com.tencent.lolm","com.riotgames.league.wildrift",
    "com.tencent.tmgp.pubgmhd","com.tencent.ig","com.tencent.tmgp.cod",
    "com.activision.callofduty.shooter","com.ea.gp.apexlegendsmobilefps",
    "com.tencent.tmgp.cf","com.dianhun.law","com.tencent.tmgp.dfm",
    "com.mojang.minecraftpe","com.minitech.miniworld","com.tencent.dawnawakening",
    "com.netease.hyxd","com.netease.mrzh","com.studiowildcard.wardrumstudios.ark",
    "zombie.survival.craft.z","com.roblox.client","com.and.games505.TerrariaPaid",
    "com.kleientertainment.doNotStarvePocket","com.miHoYo.GenshinImpact",
    "com.HoYoverse.hkrpg","com.miHoYo.bh3","com.softstar.Paladin7",
    "com.dl.dldl.aligames","com.nexon.dnf","com.blizzard.diablo.immortal",
    "com.tencent.tmgp.mumobile","com.tipsworks.android.pascalswager.google",
    "com.pearlabyss.blackdesertm.gl","com.supercell.clashofclans",
    "com.supercell.clashroyale","com.supercell.boombeach","com.netease.stzb",
    "com.aligames.sgzzlb.aligames","com.blizzard.wtcg.hearthstone",
    "com.riotgames.league.teamfighttactics","com.cygames.Shadowverse",
    "com.qihoo.tdj.aligames","com.lilithgames.ef.aligames",
    "com.tencent.tmgp.speedmobile","com.ea.game.nfs14_row","com.ea.games.r3_row",
    "com.catdaddy.nba2km","jp.konami.pesam","com.innersloth.spacemafia",
    "com.ea.game.pvz2_row","com.ustwo.monumentvalley2","com.mediascape.fallguys",
    "com.nintendo.zaka"
};

static std::unordered_map<std::string, std::string> parsePropFile(const std::string& path) {
    std::unordered_map<std::string, std::string> result;
    std::ifstream file(path); if (!file.is_open()) return result;
    std::string line;
    while (std::getline(file, line)) {
        size_t start = line.find_first_not_of(" \t");
        if (start == std::string::npos || line[start] == '#') continue;
        size_t eq = line.find('=', start); if (eq == std::string::npos) continue;
        std::string key = line.substr(start, eq - start), value = line.substr(eq + 1);
        size_t ke = key.find_last_not_of(" \t"); if (ke != std::string::npos) key = key.substr(0, ke + 1);
        size_t vs = value.find_first_not_of(" \t"), ve = value.find_last_not_of(" \t");
        value = (vs != std::string::npos) ? value.substr(vs, ve - vs + 1) : "";
        result[key] = value;
    }
    return result;
}

// JSON 解析见 common/json_config.h
using czero::parseJsonConfig;

// 进程由 timer 按检测周期拉起，跑完即退无跨次缓存
static bool g_logEnabled = false;
static void logMsg(const std::string& msg) {
    if (!g_logEnabled) return;
    time_t now = time(nullptr); struct tm tmv; localtime_r(&now, &tmv);
    char day[16]; strftime(day, sizeof(day), "%Y-%m-%d", &tmv);   // 当天日期作文件名
    std::ofstream f(LOG_DIR + day + ".log", std::ios::app); if (!f.is_open()) return;
    char ts[16]; strftime(ts, sizeof(ts), "%H:%M:%S", &tmv);
    f << "[" << ts << "] " << LOG_TAG << msg << "\n";
}

// 息屏检测
static int screenByBrightness() {   // 1亮 0灭 -1无可读节点
    std::vector<std::string> paths = {"/sys/class/leds/lcd-backlight/brightness"};
    DIR* d = opendir("/sys/class/backlight");
    if (d) {
        struct dirent* e;
        while ((e = readdir(d))) { if (e->d_name[0] == '.') continue; paths.push_back(std::string("/sys/class/backlight/") + e->d_name + "/brightness"); }
        closedir(d);
    }
    for (const auto& p : paths) { std::ifstream f(p); int b = -1; if (f.is_open()) f >> b; if (b >= 0) return b > 0 ? 1 : 0; }
    return -1;
}
static bool screenByPower() {
    FILE* p = popen("dumpsys power 2>/dev/null | grep mWakefulness= | head -1", "r"); if (!p) return true;
    char buf[128]; bool isOn = true;
    if (fgets(buf, sizeof(buf), p)) { std::string l(buf); if (l.find("Asleep") != std::string::npos || l.find("Dozing") != std::string::npos) isOn = false; }
    pclose(p); return isOn;
}
static bool isScreenOn() {
    int b = screenByBrightness();
    bool on = (b >= 0) ? (b == 1) : screenByPower();
    logMsg(on ? "屏幕亮起" : "屏幕熄灭"); return on;
}

// 前台应用检测
static bool isPkgChar(char c) {
    return (c>='a'&&c<='z') || (c>='A'&&c<='Z') || (c>='0'&&c<='9') || c=='.' || c=='_';
}
static std::string extractPackage(const std::string& line) {
    size_t slash = line.find('/');
    if (slash == std::string::npos || slash == 0) return "";
    size_t start = slash;
    while (start > 0 && isPkgChar(line[start-1])) start--;
    std::string pkg = line.substr(start, slash - start);
    return (pkg.find('.') != std::string::npos) ? pkg : "";
}
static std::string tryFocusMethod(int m) {
    const char* cmd = nullptr;
    switch (m) {
        // 取最后一条
        case 0: cmd = "dumpsys window 2>/dev/null | grep mCurrentFocus | tail -1"; break;
        case 1: cmd = "dumpsys activity activities 2>/dev/null | grep -E 'ResumedActivity' | tail -1"; break;  // 兼容 mResumedActivity 和 topResumedActivity
        case 2: cmd = "dumpsys window 2>/dev/null | grep mFocusedApp | tail -1"; break;
    }
    FILE* p = popen(cmd, "r"); if (!p) return "";
    char buf[512]; std::string result;
    if (fgets(buf, sizeof(buf), p)) result = extractPackage(std::string(buf));
    pclose(p); return result;
}
// 读 top-app cpuset，oom_score_adj==0 者即前台主进程
static std::string fgByCgroup() {
    const char* paths[] = {"/dev/cpuset/top-app/cgroup.procs", "/dev/cpuset/top-app/tasks"};
    for (const char* path : paths) {
        std::ifstream f(path); if (!f.is_open()) continue;
        std::string pid;
        while (std::getline(f, pid)) {
            if (pid.empty()) continue;
            std::ifstream adjf("/proc/" + pid + "/oom_score_adj");
            int adj = 1000; adjf >> adj;
            if (adj != 0) continue;
            std::ifstream cf("/proc/" + pid + "/cmdline");
            std::string name; std::getline(cf, name, '\0');
            size_t c = name.find(':');
            if (c != std::string::npos) name = name.substr(0, c);
            if (name.find('.') != std::string::npos) return name;
        }
    }
    return "";
}
static std::string getCurrentApp() {
    std::string r = fgByCgroup(); if (!r.empty()) return r;
    for (int i = 0; i < 3; i++) { r = tryFocusMethod(i); if (!r.empty()) return r; }  // dumpsys 兜底
    return "";
}

// 状态读一次到内存，改动累积后退出前写回一次
static std::unordered_map<std::string, std::string> g_state;
static bool g_stateDirty = false;
static void writeState() {
    std::ofstream out(STATE_FILE);
    for (const auto& kv : g_state) out << kv.first << "=" << kv.second << "\n";
}

void handleApp(const AppConfig& app, const std::string& currentApp, bool gameRunning, bool cleanEnabled) {
    bool isActive = (currentApp == app.package);
    auto it = g_state.find(app.stateKey);
    std::string state = (it != g_state.end()) ? it->second : "";
    if (isActive) {
        if (state != "active") { g_state[app.stateKey] = "active"; g_stateDirty = true; logMsg(app.name + " 活跃"); }
    } else if (!gameRunning) {
        if (state == "active") {
            // 未启用则不执行也不广播
            if (!cleanEnabled) {
                logMsg(app.name + " 清理已禁用，跳过");
            } else {
                struct stat st;
                if (stat(app.scriptPath.c_str(), &st) == 0) {
                    notifyIsland("start", app.stateKey.c_str(), 0);
                    int r = system(app.scriptPath.c_str());
                    notifyIsland("done", app.stateKey.c_str(), r == 0 ? 0 : 1);
                    logMsg("清理 " + app.name);
                }
            }
        }
        if (state != "inactive") { g_state[app.stateKey] = "inactive"; g_stateDirty = true; logMsg(app.name + " 不在前台"); }
    } else {
        logMsg("游戏运行，保持 " + app.name + " 状态");
    }
}

int main() {
    auto cfg = parseJsonConfig(CONFIG_FILE);
    g_logEnabled = (cfg["general.log"] == "true");
    if (!isScreenOn()) { logMsg("息屏，跳过"); return 0; }
    if (cfg["general.auto_clean"] != "true") { logMsg("自动清理已禁用"); return 0; }
    std::string currentApp = getCurrentApp();
    if (currentApp.empty()) { logMsg("无法获取前台应用"); return 1; }
    logMsg("前台: " + currentApp);
    bool gameRunning = (GAME_PACKAGES.count(currentApp) > 0);
    if (gameRunning) logMsg("游戏运行中");
    g_state = parsePropFile(STATE_FILE);
    for (const auto& app : MONITORED_APPS) {
        bool cleanEnabled = (cfg["app_clean." + app.stateKey + ".enabled"] == "true");
        handleApp(app, currentApp, gameRunning, cleanEnabled);
    }
    if (g_stateDirty) writeState();
    return 0;
}
