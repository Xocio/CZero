#include <iostream>
#include <fstream>
#include <string>
#include <unordered_map>
#include <unordered_set>
#include <vector>
#include <cstdlib>
#include <cstdio>
#include <ctime>
#include <unistd.h>
#include <csignal>
#include <dirent.h>
#include <iterator>
#include "common/json_config.h"
#include "common/file_lock.h"
const std::string MODULE_PATH="/data/adb/modules/CZero/";
const std::string CONFIG_FILE=MODULE_PATH+"config.json";
const std::string COUNT_FILE=MODULE_PATH+"basis/basis.prop";
const std::string LOG_DIR=MODULE_PATH+"log/";
const std::string LOG_TAG="[压制] ";
const std::string LOCK_FILE=MODULE_PATH+"list/suppress/lock";
// JSON 解析见 common/json_config.h
using czero::parseJsonConfig;
// 进程跑完即退，无需任何跨次缓存
static bool g_logEnabled=false;
static void logMsg(const std::string& msg){
    if(!g_logEnabled) return;
    time_t now=time(nullptr); struct tm tmv; localtime_r(&now,&tmv);
    char day[16]; strftime(day,sizeof(day),"%Y-%m-%d",&tmv);   // 当天日期作文件名
    std::ofstream f(LOG_DIR+day+".log",std::ios::app); if(!f.is_open()) return;
    char ts[16]; strftime(ts,sizeof(ts),"%H:%M:%S",&tmv);
    f<<"["<<ts<<"] "<<LOG_TAG<<msg<<"\n";
}
// 息屏检测
static int screenByBrightness(){   // 1亮 0灭 -1无可读节点
    std::vector<std::string> paths={"/sys/class/leds/lcd-backlight/brightness"};
    DIR* d=opendir("/sys/class/backlight");
    if(d){struct dirent* e; while((e=readdir(d))){if(e->d_name[0]=='.') continue; paths.push_back(std::string("/sys/class/backlight/")+e->d_name+"/brightness");} closedir(d);}
    for(const auto& p:paths){std::ifstream f(p); int b=-1; if(f.is_open()) f>>b; if(b>=0) return b>0?1:0;}
    return -1;
}
static bool screenByPower(){
    FILE* p=popen("dumpsys power 2>/dev/null | grep mWakefulness= | head -1","r"); if(!p) return true;
    char buf[128]; bool isOn=true;
    if(fgets(buf,sizeof(buf),p)){std::string l(buf); if(l.find("Asleep")!=std::string::npos||l.find("Dozing")!=std::string::npos) isOn=false;}
    pclose(p); return isOn;
}
static bool isScreenOn(){
    int b=screenByBrightness();
    return (b>=0)?(b==1):screenByPower();
}
// 前台应用检测
static bool isPkgChar(char c){
    return (c>='a'&&c<='z')||(c>='A'&&c<='Z')||(c>='0'&&c<='9')||c=='.'||c=='_';
}
static std::string extractPackage(const std::string& line){
    size_t slash=line.find('/');
    if(slash==std::string::npos||slash==0) return "";
    size_t start=slash;
    while(start>0&&isPkgChar(line[start-1])) start--;
    std::string pkg=line.substr(start,slash-start);
    return (pkg.find('.')!=std::string::npos)?pkg:""; 
}
static std::string tryFocusMethod(int m){
    const char* cmd=nullptr;
    switch(m){
        case 0: cmd="dumpsys window 2>/dev/null | grep mCurrentFocus | tail -1"; break;
        case 1: cmd="dumpsys activity activities 2>/dev/null | grep -E 'ResumedActivity' | tail -1"; break;  // 兼容 mResumedActivity 和 topResumedActivity
        case 2: cmd="dumpsys window 2>/dev/null | grep mFocusedApp | tail -1"; break;
    }
    FILE* p=popen(cmd,"r"); if(!p) return "";
    char buf[512]; std::string result;
    if(fgets(buf,sizeof(buf),p)) result=extractPackage(std::string(buf));
    pclose(p); return result;
}
// 读 top-app cpuset，oom_score_adj==0 者即前台主进程
static std::string fgByCgroup(){
    const char* paths[]={"/dev/cpuset/top-app/cgroup.procs","/dev/cpuset/top-app/tasks"};
    for(const char* path:paths){
        std::ifstream f(path); if(!f.is_open()) continue;
        std::string pid;
        while(std::getline(f,pid)){
            if(pid.empty()) continue;
            std::ifstream adjf("/proc/"+pid+"/oom_score_adj");
            int adj=1000; adjf>>adj;
            if(adj!=0) continue;
            std::ifstream cf("/proc/"+pid+"/cmdline");
            std::string name; std::getline(cf,name,'\0');
            size_t c=name.find(':');
            if(c!=std::string::npos) name=name.substr(0,c);
            if(name.find('.')!=std::string::npos) return name;
        }
    }
    return "";
}
static std::string getCurrentApp(){
    std::string r=fgByCgroup(); if(!r.empty()) return r;
    for(int i=0;i<3;i++){r=tryFocusMethod(i); if(!r.empty()) return r;}  // dumpsys 兜底
    return "";
}
// 重建完整进程名

std::string fullProcName(const std::string& package, const std::string& p){
    size_t c=p.find(':');
    return (c!=std::string::npos)?package+":"+p.substr(c+1):package+":"+p;
}
// 扫描 /proc 一次，按完整进程名精确匹配
std::vector<std::string> killByNames(const std::unordered_set<std::string>& targets){
    std::vector<std::string> killed;
    if(targets.empty()) return killed;
    DIR* d=opendir("/proc"); if(!d) return killed;
    struct dirent* e;
    while((e=readdir(d))){
        pid_t pid=atoi(e->d_name);
        if(pid<=0) continue;
        std::ifstream f("/proc/"+std::string(e->d_name)+"/cmdline");
        if(!f.is_open()) continue;
        std::string name; std::getline(f,name,'\0');  // cmdline 以 NUL 分隔，首段即进程名
        if(name.empty()) continue;
        if(targets.count(name)&&kill(pid,SIGKILL)==0) killed.push_back(name+"("+std::to_string(pid)+")");
    }
    closedir(d);
    return killed;
}
void updateStatistics(int killed){
    czero::FileLock lock(czero::kBasisLock);  // 统计写入互斥
    std::unordered_map<std::string,int> counts;
    std::ifstream in(COUNT_FILE);
    if(in.is_open()){std::string line; while(std::getline(in,line)){size_t pos=line.find('='); if(pos!=std::string::npos){try{counts[line.substr(0,pos)]=std::stoi(line.substr(pos+1));}catch(...){}}} in.close();}
    counts["suppress"]+=killed;
    std::ofstream out(COUNT_FILE); for(const auto& kv:counts) out<<kv.first<<"="<<kv.second<<"\n";
}
bool tryLock(){
    std::ifstream in(LOCK_FILE);
    if(in.is_open()){
        std::string pid;
        if(std::getline(in,pid)&&!pid.empty()){
            // 读 /proc/<pid>/cmdline 确认占锁进程仍是 suppress
            std::ifstream cf("/proc/"+pid+"/cmdline");
            std::string name; if(cf.is_open()) std::getline(cf,name,'\0');
            if(name.find("suppress")!=std::string::npos) return false;
        }
    }
    std::ofstream out(LOCK_FILE); if(out.is_open()){out<<getpid()<<"\n"; return true;} return false;
}
void unlock(){remove(LOCK_FILE.c_str());}
void signalHandler(int sig){unlock();exit(sig);}
struct AppProcesses{std::string package;std::vector<std::string> processes;std::string name;};
const std::vector<AppProcesses> APP_CONFIGS={
    {"com.tencent.mm",{"mm:toolsmp","mm:appbrand1","mm:support","mm:tools","mm:appbrand2","mm:appbrand0","mm:jectl","mm:hotpot","mm:xweb"},"微信"},
    {"com.tencent.mobileqq",{"mobileqq:mini3","mobileqq:mini","mobileqq:qzone","mobileqq:tool"},"QQ"},
    {"com.eg.android.AlipayGphone",{"AlipayGphone:sandboxed_privilege_process0","AlipayGphone:gpu_process","AlipayGphone:widgetProvider","AlipayGphone:lite2"},"支付宝"}
};
int main(int argc, char* argv[]){
    signal(SIGINT,signalHandler); signal(SIGTERM,signalHandler);
    // force手动执行时无视 enabled 开关
    bool force = (argc > 1 && std::string(argv[1]) == "force");
    auto cfg=parseJsonConfig(CONFIG_FILE);
    if(!force && cfg["suppress.enabled"]!="true") return 0;
    g_logEnabled=(cfg["general.log"]=="true");
    if(!tryLock()) return 0;
    if(!isScreenOn()){logMsg("息屏，跳过");unlock();return 0;}
    std::string foregroundApp=getCurrentApp();
    std::string fgName;
    for(const auto& app:APP_CONFIGS) if(app.package==foregroundApp){fgName=app.name; break;}
    if(foregroundApp.empty()) logMsg("前台: (未识别，本轮不压制)");
    else logMsg("前台: "+foregroundApp+(fgName.empty()?"":" ("+fgName+")"));
    // 前台 App 不动，其余 App 的后台进程加入待杀集合
    std::unordered_set<std::string> targets;
    std::vector<std::string> scanApps;
    for(const auto& app:APP_CONFIGS){
        if(foregroundApp.empty()||foregroundApp==app.package) continue;
        for(const auto& p:app.processes) targets.insert(fullProcName(app.package,p));
        scanApps.push_back(app.name);
    }
    if(!scanApps.empty()){
        std::string s; for(size_t i=0;i<scanApps.size();i++){if(i)s+="、"; s+=scanApps[i];}
        logMsg("待压制目标: "+s+"，共 "+std::to_string(targets.size())+" 个进程名");
    }
    std::vector<std::string> killed=killByNames(targets);
    if(!killed.empty()){
        updateStatistics((int)killed.size());
        std::string s; for(size_t i=0;i<killed.size();i++){if(i)s+="、"; s+=killed[i];}
        logMsg("已杀 "+std::to_string(killed.size())+" 个后台进程: "+s);
    }else{
        logMsg("无可杀进程");
    }
    logMsg("完成"); unlock(); return 0;
}
