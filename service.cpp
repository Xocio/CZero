#include <iostream>
#include <fstream>
#include <string>
#include <vector>
#include <cstdlib>
#include <sys/stat.h>

const std::string MODDIR = "/data/adb/modules/CZero";
const std::string MODULE_PROP = MODDIR + "/module.prop";
const std::string LOG_DIR = MODDIR + "/log";
const std::string BASIS_DIR = MODDIR + "/basis";
const std::string BASIS_PROP = BASIS_DIR + "/basis.prop";

bool file_exists(const std::string& path) {
    struct stat buf;
    return (stat(path.c_str(), &buf) == 0);
}

void update_module_desc() {
    if (!file_exists(MODULE_PROP)) return;

    std::ifstream in(MODULE_PROP);
    std::vector<std::string> lines;
    std::string line;

    while (std::getline(in, line)) {
        if (line.find("description=") == 0) {
            line = "description=✅ The module is active";
        } else if (line.find("descriptionAnsi=") == 0) {
            line = R"(descriptionAnsi=\e[1;36mCZero\e[0m丨\e[33mThe module is active)";
        }
        lines.push_back(line);
    }
    in.close();

    std::ofstream out(MODULE_PROP);
    for (const auto& l : lines) {
        out << l << std::endl;
    }
}

int main() {
    for (const std::string& dir : {LOG_DIR, BASIS_DIR, MODDIR + "/cron"})
        mkdir(dir.c_str(), 0777);

    if (!file_exists(BASIS_PROP)) {
        std::ofstream(BASIS_PROP).close();
    }

    update_module_desc();

    std::ofstream(MODDIR + "/status") << "模块配置完成" << std::endl;
    return 0;
}
