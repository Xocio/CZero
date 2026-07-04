#include <iostream>
#include <fstream>
#include <string>
#include <chrono>
#include <thread>
#include <cstdlib>
#include <iomanip>
#include <sstream>
#include <filesystem>
#include <map>
#include <vector>
#include <cstdio>
#include <cerrno>
#include <cstring>
#include <cctype>
#include <unordered_map>
#include <iterator>
#include <functional>
#include "common/json_config.h"
#include "common/file_lock.h"

// 通知

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

// JSON 解析见 common/json_config.h
using czero::parseJsonConfig;

class F2FSCleaner {
private:
    // 配置文件路径
    const std::string CONFIG_FILE = "/data/adb/modules/CZero/config.json";
    const std::string LOG_DIR = "/data/adb/modules/CZero/log";
    const std::string LOG_TAG = "[GC] ";
    const std::string BASIS_FILE = "/data/adb/modules/CZero/basis/basis.prop";

    // 全局变量
    int INITIAL_DIRTY = 0;
    int FINAL_DIRTY = 0;
    int CLEANED_BLOCKS = 0;
    int DIRTY_THRESHOLD = 800;   // 默认触发阈值
    int CLEAN_THRESHOLD = 300;   // 默认完成阈值
    int WAIT_SCREEN_OFF_TIMEOUT = 30; // 默认等待息屏超时
    int MAX_RUNTIME_SEC = 600;        // 默认最大运行时

    bool LOG_ENABLED = false;
    bool GC_ENABLED = false;

    std::string TARGET_DEVICE;

private:
    // 获取当前时间
    std::string getCurrentTime() {
        auto now = std::chrono::system_clock::now();
        auto time_t = std::chrono::system_clock::to_time_t(now);
        std::stringstream ss;
        ss << std::put_time(std::localtime(&time_t), "%Y-%m-%d %H:%M:%S");
        return ss.str();
    }

    // 当天日期作日志文件名
    std::string logPath() {
        auto now = std::chrono::system_clock::now();
        auto time_t = std::chrono::system_clock::to_time_t(now);
        std::stringstream ss;
        ss << LOG_DIR << "/" << std::put_time(std::localtime(&time_t), "%Y-%m-%d") << ".log";
        return ss.str();
    }

    // 日志输出函数
    void logOutput(const std::string& message) {
        if (LOG_ENABLED) {
            std::ofstream logFile(logPath(), std::ios::app);
            if (logFile.is_open()) {
                logFile << "[" << getCurrentTime() << "] " << LOG_TAG << message << std::endl;
            }
        }
    }

    // 显示函数
    void display(const std::string& message) {
        std::cout << "[" << getCurrentTime() << "] " << message << std::endl;
        logOutput(message);
    }

    // 去除字符串空格
    std::string trim(const std::string& str) {
        if (str.empty()) return "";
        size_t first = 0;
        while (first < str.size() && std::isspace(static_cast<unsigned char>(str[first]))) ++first;
        if (first == str.size()) return "";
        size_t last = str.size();
        while (last > first && std::isspace(static_cast<unsigned char>(str[last - 1]))) --last;
        return str.substr(first, last - first);
    }

    // 读取配置文件
    bool readConfig() {
        auto cfg = parseJsonConfig(CONFIG_FILE);
        if (cfg.empty()) {
            logOutput("配置文件不存在或为空: " + CONFIG_FILE);
            return false;
        }

        GC_ENABLED = (cfg["gc.enabled"] == "true");
        LOG_ENABLED = (cfg["general.log"] == "true");

        auto readInt = [&](const std::string& key, int minVal, int& target) {
            auto it = cfg.find(key);
            if (it == cfg.end() || it->second.empty()) return;
            try { int v = std::stoi(it->second); if (v >= minVal) target = v; }
            catch (...) { logOutput(key + " 解析失败: '" + it->second + "'"); }
        };
        readInt("gc.dirty_threshold", 1, DIRTY_THRESHOLD);
        readInt("gc.clean_threshold", 1, CLEAN_THRESHOLD);
        readInt("gc.wait_screen_off_timeout", 5, WAIT_SCREEN_OFF_TIMEOUT);
        readInt("gc.max_runtime_sec", 60, MAX_RUNTIME_SEC);

        // 阈值关系校验与矫正
        if (CLEAN_THRESHOLD >= DIRTY_THRESHOLD) {
            int oldDirty = DIRTY_THRESHOLD;
            DIRTY_THRESHOLD = CLEAN_THRESHOLD + 50; // 保证触发阈值高于完成阈值
            logOutput("阈值矫正: CLEAN_THRESHOLD(" + std::to_string(CLEAN_THRESHOLD) +
                      ") >= DIRTY_THRESHOLD(" + std::to_string(oldDirty) + "), 将 DIRTY_THRESHOLD 调整为 " + std::to_string(DIRTY_THRESHOLD));
        }
        return true;
    }

    // 初始化日志目录
    void initLog() {
        if (LOG_ENABLED) {
            std::error_code ec;
            std::filesystem::create_directories(LOG_DIR, ec);
            if (ec) {
                std::cerr << "创建日志目录失败: " << LOG_DIR << " (" << ec.message() << ")" << std::endl;
                return;
            }
        }
    }

    // 读取文件内容
    std::string readFile(const std::string& filepath) {
        std::ifstream file(filepath);
        if (!file.is_open()) {
            return "";
        }

        std::string content;
        std::getline(file, content);
        file.close();
        return trim(content);
    }

    // 写入文件
    bool writeFile(const std::string& filepath, const std::string& content) {
        std::ofstream file(filepath);
        if (!file.is_open()) {
            logOutput("写入失败: " + filepath + ", errno=" + std::to_string(errno) + ", " + std::string(std::strerror(errno)));
            return false;
        }
        file << content;
        return true;
    }

    // 获取设备名
    std::string getDeviceName(const std::string& devicePath) {
        size_t pos = devicePath.find_last_of('/');
        if (pos != std::string::npos) {
            return devicePath.substr(pos + 1);
        }
        return devicePath;
    }

    bool isValidF2fsDevice(const std::string& dev) {
        std::string base = getDeviceName(dev);
        std::string p = "/sys/fs/f2fs/" + base + "/dirty_segments";
        return std::filesystem::exists(p);
    }

    // 获取/data分区的实际设备
    std::string getDataDevice() {
        // 通过挂载点查找并校验 sysfs
        {
            std::ifstream mounts("/proc/mounts");
            if (mounts.is_open()) {
                std::string line;
                while (std::getline(mounts, line)) {
                    if (line.find(" /data ") != std::string::npos) {
                        std::istringstream iss(line);
                        std::string device, on, mountpoint, fstype;
                        iss >> device >> mountpoint >> fstype;
                        std::string base = getDeviceName(device);
                        if (isValidF2fsDevice(base)) return base;
                    }
                }
            }
        }

        // 通过系统属性
        {
            std::string propResult = readFile("/data/property/persist.sys.mnt.dev.data");
            if (!propResult.empty()) {
                std::string base = getDeviceName(propResult);
                if (isValidF2fsDevice(base)) return base;
            }
        }

        // 扫描 /sys/fs/f2fs 下目录并比对是否挂载 /data
        try {
            std::string f2fsPath = "/sys/fs/f2fs/";
            for (const auto& entry : std::filesystem::directory_iterator(f2fsPath)) {
                if (!entry.is_directory()) continue;
                std::string deviceName = entry.path().filename();
                if (deviceName == "features") continue;
                std::ifstream mounts("/proc/mounts");
                std::string line;
                while (std::getline(mounts, line)) {
                    if (line.find(deviceName) != std::string::npos && line.find("/data") != std::string::npos) {
                        return deviceName;
                    }
                }
            }
        } catch (...) {}

        // 兜底
        try {
            std::vector<std::string> candidates;
            for (const auto& entry : std::filesystem::directory_iterator("/sys/fs/f2fs/")) {
                if (!entry.is_directory()) continue;
                std::string name = entry.path().filename();
                if (name == "features") continue;
                std::string base = "/sys/fs/f2fs/" + name;
                if (std::filesystem::exists(base + "/gc_urgent") &&
                    std::filesystem::exists(base + "/dirty_segments")) {
                    candidates.push_back(name);
                }
            }
            if (candidates.size() == 1) {
                logOutput("未能关联 /data，退化选用唯一 f2fs 节点: " + candidates[0]);
                return candidates[0];
            }
            if (candidates.size() > 1) {
                logOutput("未能关联 /data 且存在多个 f2fs 节点，取第一个: " + candidates[0]);
                return candidates[0];
            }
        } catch (...) {}

        return "";
    }

    // 获取段信息
    struct SegmentInfo {
        int dirty;
        int free;
        int used;
        int total;
    };

    SegmentInfo getSegmentInfo(const std::string& device) {
        SegmentInfo info {0,0,0,0};
        std::string sysfsPath = "/sys/fs/f2fs/" + device;

        std::string dirtyStr = readFile(sysfsPath + "/dirty_segments");
        std::string freeStr  = readFile(sysfsPath + "/free_segments");

        try { if (!dirtyStr.empty()) info.dirty = std::stoi(dirtyStr); } catch (...) { logOutput("dirty_segments 解析失败: '" + dirtyStr + "'"); }
        try { if (!freeStr.empty())  info.free  = std::stoi(freeStr); } catch (...) { logOutput("free_segments 解析失败: '" + freeStr + "'"); }

        std::string totalStr = readFile(sysfsPath + "/segment_count");
        if (totalStr.empty()) totalStr = readFile(sysfsPath + "/main_segments");
        try {
            if (!totalStr.empty()) {
                info.total = std::stoi(totalStr);
                info.used = info.total - info.free - info.dirty;
                if (info.used < 0) info.used = 0;
            }
        } catch (...) { logOutput("segment_count/main_segments 解析失败: '" + totalStr + "'"); }

        return info;
    }

    // 获取设备信息
    void getDeviceInfo(const std::string& device) {
        SegmentInfo info = getSegmentInfo(device);
        display("当前F2FS挂载设备: " + device);
        if (info.dirty > 0) display("脏段: " + std::to_string(info.dirty));
        if (info.free  > 0) display("空闲段: " + std::to_string(info.free));
    }


    // 监控GC进度
    void monitorGcProgress(const std::string& device) {
        const int MAX_RUNTIME = MAX_RUNTIME_SEC; // 使用配置
        const int CHECK_INTERVAL = 5; // 检查间隔

        auto startTime = std::chrono::steady_clock::now();
        int lastDirty = -1;

        display("开始监控GC");
        display("触发阈值: " + std::to_string(DIRTY_THRESHOLD) + ", 完成目标: " + std::to_string(CLEAN_THRESHOLD));

        while (true) {
            auto currentTime = std::chrono::steady_clock::now();
            auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(currentTime - startTime).count();

            if (elapsed >= MAX_RUNTIME) {
                display("已达到最大运行时间（" + std::to_string(MAX_RUNTIME/60) + "分钟），停止监控");
                break;
            }

            SegmentInfo info = getSegmentInfo(device);
            logOutput("运行: " + std::to_string(elapsed) + "s | 脏段: " + std::to_string(info.dirty) + " | 空闲段: " + std::to_string(info.free));

            if (info.dirty <= CLEAN_THRESHOLD) {
                display("脏段数量已降至 " + std::to_string(info.dirty) + "，达到完成目标");
                break;
            }

            if (info.dirty != lastDirty) {
                lastDirty = info.dirty;
            }

            std::this_thread::sleep_for(std::chrono::seconds(CHECK_INTERVAL));
        }
    }

    // 记录初始状态
    void recordInitialState(const std::string& device) {
        std::string dirtyStr = readFile("/sys/fs/f2fs/" + device + "/dirty_segments");
        try { INITIAL_DIRTY = dirtyStr.empty() ? 0 : std::stoi(dirtyStr); }
        catch (...) { INITIAL_DIRTY = 0; logOutput("读取初始脏段失败: '" + dirtyStr + "'"); }
        logOutput("初始脏段: " + std::to_string(INITIAL_DIRTY));
    }

    // 记录最终状态并计算清理量
    void recordFinalState(const std::string& device) {
        std::string dirtyStr = readFile("/sys/fs/f2fs/" + device + "/dirty_segments");
        try { FINAL_DIRTY = dirtyStr.empty() ? 0 : std::stoi(dirtyStr); }
        catch (...) { FINAL_DIRTY = 0; logOutput("读取最终脏段失败: '" + dirtyStr + "'"); }
        CLEANED_BLOCKS = INITIAL_DIRTY - FINAL_DIRTY; if (CLEANED_BLOCKS < 0) CLEANED_BLOCKS = 0;
        logOutput("最终脏段: " + std::to_string(FINAL_DIRTY) + ", 清理: " + std::to_string(CLEANED_BLOCKS));
    }

    // 检查是否需要执行GC
    bool checkGcNeeded(const std::string& device) {
        std::string dirtyStr = readFile("/sys/fs/f2fs/" + device + "/dirty_segments");
        int currentDirty = 0;
        try { currentDirty = dirtyStr.empty() ? 0 : std::stoi(dirtyStr); }
        catch (...) { logOutput("当前脏段解析失败: '" + dirtyStr + "'"); }

        logOutput("当前脏段: " + std::to_string(currentDirty) + ", 触发阈值: " + std::to_string(DIRTY_THRESHOLD));
        return currentDirty > DIRTY_THRESHOLD;
    }

    // 息屏检测
    bool isScreenOff() {
        logOutput("开始息屏检测...");

        // 扫 /sys/class/backlight 下所有背光节点
        {
            std::vector<std::string> backlightPaths = {
                "/sys/class/leds/lcd-backlight/brightness"
            };
            std::error_code ec;
            for (const auto& entry : std::filesystem::directory_iterator("/sys/class/backlight", ec)) {
                backlightPaths.push_back(entry.path().string() + "/brightness");
            }
            for (const auto& path : backlightPaths) {
                std::string brightness = readFile(path);
                if (!brightness.empty()) {
                    try {
                        int value = std::stoi(brightness);
                        logOutput("背光亮度(" + path + "): " + std::to_string(value));
                        if (value == 0)  { logOutput("检测到息屏状态"); return true; }
                        if (value > 0)   { logOutput("检测到亮屏状态"); return false; }
                    } catch (...) { logOutput("背光值解析失败: '" + brightness + "' from " + path); }
                }
            }
        }

        // dumpsys power 兜底
        {
            std::string powerCmd = "timeout 3 dumpsys power 2>/dev/null";
            FILE* powerPipe = popen(powerCmd.c_str(), "r");
            if (powerPipe != nullptr) {
                std::string result; char buffer[1024];
                while (fgets(buffer, sizeof(buffer), powerPipe) != nullptr) { result += buffer; }
                int ret = pclose(powerPipe); (void)ret;
                std::string s = trim(result);
                if (!s.empty()) {
                    logOutput("电源状态: " + (s.size()>160 ? s.substr(0,160)+"..." : s));
                    if (s.find("mWakefulness=Asleep") != std::string::npos ||
                        s.find("Display Power: state=OFF") != std::string::npos ||
                        s.find("mInteractive=false") != std::string::npos) {
                        logOutput("screenOff==true");
                        return true;
                    }
                    if (s.find("mWakefulness=Awake") != std::string::npos ||
                        s.find("Display Power: state=ON") != std::string::npos ||
                        s.find("mInteractive=true") != std::string::npos) {
                        logOutput("screenOff==false");
                        return false;
                    }
                } else {
                    logOutput("dumpsys power 返回空");
                }
            } else {
                logOutput("无法执行 dumpsys power");
            }
        }

        // 如果仍无法判断，默认认为屏幕开启
        logOutput("无法确定屏幕状态，默认认为屏幕开启");
        return false;
    }

    // 等待息屏状态
    bool waitForScreenOff() {
        logOutput("开始等待息屏状态，最长等待" + std::to_string(WAIT_SCREEN_OFF_TIMEOUT) + "秒...");
        auto startTime = std::chrono::steady_clock::now();
        while (true) {
            auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(std::chrono::steady_clock::now() - startTime).count();
            if (elapsed >= WAIT_SCREEN_OFF_TIMEOUT) {
                logOutput("等待超时（" + std::to_string(WAIT_SCREEN_OFF_TIMEOUT) + "秒），退出执行");
                return false;
            }
            if (isScreenOff()) {
                logOutput("检测到息屏状态，继续执行清理（等待了" + std::to_string((int)elapsed) + "秒）");
                return true;
            }
            display("屏幕仍然开启，继续等待... (剩余" + std::to_string(WAIT_SCREEN_OFF_TIMEOUT - (int)elapsed) + "秒)");
            std::this_thread::sleep_for(std::chrono::seconds(3));
        }
    }

    // 触发 GC
    bool triggerGcUrgent(const std::string& gcUrgentPath) {
        for (int i = 0; i < 3; ++i) {
            if (writeFile(gcUrgentPath, "1")) return true;
            std::this_thread::sleep_for(std::chrono::milliseconds(200));
        }
        logOutput("错误: 无法写入 " + gcUrgentPath + "，检测权限");
        return false;
    }

public:
    ~F2FSCleaner() = default;

    // 主执行函数
    int run(bool force = false) {
        if (!readConfig()) {
            logOutput("错误: 无法读取配置文件");
            return 1;
        }

        initLog();

        if (!GC_ENABLED && !force) {
            logOutput("GC脏段清理功能已禁用，退出");
            return 0;
        }

        if (force) {
            logOutput("手动执行，跳过息屏等待");
        } else if (!waitForScreenOff()) {
            logOutput("等待息屏超时，程序退出");
            return 0;
        }

        TARGET_DEVICE = getDataDevice();
        if (TARGET_DEVICE.empty()) {
            logOutput("错误: 未找到可用的 F2FS 设备，你的设备不支持");
            return 1;
        }

        logOutput("找到设备: " + TARGET_DEVICE);

        recordInitialState(TARGET_DEVICE);


        if (!checkGcNeeded(TARGET_DEVICE)) {

            recordFinalState(TARGET_DEVICE);

            updateGccleanStats(CLEANED_BLOCKS);
            logOutput("检查完成，无需清理");
            return 0;
        }

        notifyIsland("start", "gc", 0);
        auto finish = [](int rc) -> int { notifyIsland("done", "gc", rc); return rc; };

        display("设备信息:");
        getDeviceInfo(TARGET_DEVICE);

        std::string sysfsPath = "/sys/fs/f2fs/" + TARGET_DEVICE;
        std::string gcRemainingPath = sysfsPath + "/gc_remaining_segments";
        std::string gcUrgentPath    = sysfsPath + "/gc_urgent";

        // gc_urgent复位为 0
        struct ScopeExit {
            std::function<void()> f; bool armed = false;
            ~ScopeExit() { if (armed && f) f(); }
        } gcReset{ [this, gcUrgentPath]() { writeFile(gcUrgentPath, "0"); logOutput("gc_urgent 已复位为 0"); } };

        auto canReadInt = [&](const std::string& path, int& out) -> bool {
            std::string s = readFile(path);
            if (s.empty()) return false;
            try { out = std::stoi(s); return out >= 0; } catch (...) { logOutput("整数解析失败: '" + s + "' from " + path); return false; }
        };

        int rem = -1;
        bool hasProgress = std::filesystem::exists(gcRemainingPath) && canReadInt(gcRemainingPath, rem);

        if (!hasProgress) {
            display("正在触发GC回收...");
            if (std::filesystem::exists(gcUrgentPath)) {
                if (!triggerGcUrgent(gcUrgentPath)) { return finish(1); }
                gcReset.armed = true;
                display("GC已触发成功");
            } else {
                logOutput("错误: 找不到gc_urgent接口，无法触发GC");
                return finish(1);
            }

            monitorGcProgress(TARGET_DEVICE);

            recordFinalState(TARGET_DEVICE);

            display("GC完成");
            getDeviceInfo(TARGET_DEVICE);
        } else {
            // 进度显示的情况
            int initial = rem;
            if (initial == 0) {
                display("无需GC: 剩余可回收段数已经是 0");
                recordFinalState(TARGET_DEVICE);
                updateGccleanStats(CLEANED_BLOCKS);
                return finish(0);
            }

            display("触发GC回收 (初始段数: " + std::to_string(initial) + ")...");
            if (!triggerGcUrgent(gcUrgentPath)) { return finish(1); }
            gcReset.armed = true;

            // 进度监控
            display("回收进度:");
            auto startTime = std::chrono::steady_clock::now();
            while (true) {
                int current = 0;
                if (!canReadInt(gcRemainingPath, current)) { logOutput("错误: 无法读取当前段数"); break; }
                int processed = initial - current;
                int progress = (initial > 0) ? (processed * 100 / initial) : 100;
                auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(std::chrono::steady_clock::now() - startTime).count();

                logOutput("进度: " + std::to_string(progress) + "% [" + std::to_string(processed) + "/" + std::to_string(initial) + "] | 用时: " + std::to_string(elapsed) + "s");

                if (current == 0) break;
                // 超时保护，GC 停滞时退出
                if (elapsed >= MAX_RUNTIME_SEC) {
                    display("已达到最大运行时间（" + std::to_string(MAX_RUNTIME_SEC / 60) + "分钟），停止等待");
                    break;
                }
                std::this_thread::sleep_for(std::chrono::seconds(1));
            }
            recordFinalState(TARGET_DEVICE);

            display("GC回收完成");
            getDeviceInfo(TARGET_DEVICE);
        }

        updateGccleanStats(CLEANED_BLOCKS);

        logOutput("清理任务完成，程序正常退出");
        return finish(0);
    }

    // 更新basis.prop文件中的gcclean累计值
    void updateGccleanStats(int cleanedCount) {
        if (cleanedCount <= 0) return;

        std::string basisDir = "/data/adb/modules/CZero/basis";
        std::filesystem::create_directories(basisDir);

        double increment = static_cast<double>(cleanedCount) / 1000.0;

        // 统计写入互斥
        czero::FileLock lock(czero::kBasisLock);

        // 读一次即拿到全部键值，gcclean 累加值也从其中取
        std::map<std::string, std::string> props;
        std::ifstream basisIn(BASIS_FILE);
        if (basisIn.is_open()) {
            std::string line;
            while (std::getline(basisIn, line)) {
                size_t pos = line.find('=');
                if (pos != std::string::npos) {
                    props[line.substr(0, pos)] = line.substr(pos + 1);
                }
            }
            basisIn.close();
        }

        double currentValue = 0.0;
        auto it = props.find("gcclean");
        if (it != props.end()) { try { currentValue = std::stod(it->second); } catch (...) {} }
        double newValue = currentValue + increment;

        std::ostringstream oss; oss << std::fixed << std::setprecision(2) << newValue;
        props["gcclean"] = oss.str();

        std::ofstream basisOut(BASIS_FILE);
        if (basisOut.is_open()) {
            for (const auto& prop : props) {
                basisOut << prop.first << "=" << prop.second << std::endl;
            }
        }

        std::ostringstream logMsg;
        logMsg << "累计统计: +" << std::fixed << std::setprecision(2) << increment
               << " (总计: " << std::fixed << std::setprecision(2) << newValue << ")";
        logOutput(logMsg.str());
    }
};

// 主函数
int main(int argc, char* argv[]) {
    bool force = (argc > 1 && std::string(argv[1]) == "force");
    F2FSCleaner cleaner;
    return cleaner.run(force);
}