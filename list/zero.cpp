#include <fstream>
#include <string>
#include <ctime>
#include <dirent.h>
#include <cstdio>
#include "common/file_lock.h"

//清理非今天的旧日志文件
static void cleanOldLogs() {
    const std::string logDir = "/data/adb/modules/CZero/log";
    time_t now = time(nullptr); char today[16];
    strftime(today, sizeof(today), "%Y-%m-%d", localtime(&now));
    std::string keep = std::string(today) + ".log";   // 仅保留今天的日志
    DIR* d = opendir(logDir.c_str());
    if (!d) return;
    struct dirent* e;
    while ((e = readdir(d))) {
        std::string name = e->d_name;
        if (name.size() <= 4 || name.substr(name.size() - 4) != ".log") continue;
        if (name == keep) continue;
        remove((logDir + "/" + name).c_str());
    }
    closedir(d);
}

int main() {
    cleanOldLogs();

    const std::string filePath = "/data/adb/modules/CZero/basis/basis.prop";

    czero::FileLock lock(czero::kBasisLock);  // 统计写入互斥
    std::ifstream inFile(filePath);
    if (!inFile) {
        return 1;
    }

    std::string line;
    std::string fileContent;
    while (std::getline(inFile, line)) {
        if (line.rfind("statistics_today=", 0) == 0) {
            line = "statistics_today=0";
        }
        fileContent += line + "\n";
    }
    inFile.close();

    std::ofstream outFile(filePath);
    if (!outFile) {
        return 1;
    }
    outFile << fileContent;
    outFile.close();
    return 0;
}
