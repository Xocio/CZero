#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/timerfd.h>
#include <sys/epoll.h>
#include <sys/stat.h>
#include <sys/file.h>
#include <stdint.h>
#include <string.h>
#include <time.h>
#include <signal.h>
#include <errno.h>
#include <stdarg.h>
#include <spawn.h>
#include <sys/wait.h>
#include <fcntl.h>
#include <string>
#include <unordered_map>
#include "common/json_config.h"

#define MAX_JOBS 100
#define MAX_LINE 1024
#define CONFIG_FILE "/data/adb/modules/CZero/config.json"
#define MODDIR_S "/data/adb/modules/CZero"
#define LOG_DIR "/data/adb/modules/CZero/log/"
#define LOG_TAG "[定时] "
#define LOCK_FILE "/data/adb/modules/CZero/cron/timer.lock"
#define STATE_FILE "/data/adb/modules/CZero/cron/state"
#define MAX_GATES 32

extern char **environ;

typedef struct {
    int minute[60];
    int hour[24];
    int day[32];
    int month[13];
    int weekday[7];
    int day_spec;
    int weekday_spec;
    char cmd[512];
    pid_t pid;
} Job;

// 门控，周期大于或者等于1天的任务，由 load_config_jobs 从 config.json 计算
typedef struct { char cmd[512]; long min_sec; } Gate;
// 运行状态，持久化每个门控任务的上次运行时间戳，存于 cron/state
typedef struct { char cmd[512]; time_t last_run; } RunState;

static volatile sig_atomic_t running = 1;
static volatile sig_atomic_t force_reload = 0;
static bool g_logEnabled = false;
static Job jobs[MAX_JOBS];
static int njobs = 0;
static time_t mtime = 0;
static int lockfd = -1;
static Gate gates[MAX_GATES];
static int ngates = 0;
static RunState states[MAX_GATES];
static int nstates = 0;

// 取当前时间
static void get_local_time(struct tm *tm_out) {
    time_t now = time(NULL);
    localtime_r(&now, tm_out);
}

void write_log(const char *fmt, ...) {
    if (!g_logEnabled) return;

    struct tm t;
    get_local_time(&t);

    char day[16];
    strftime(day, sizeof(day), "%Y-%m-%d", &t);   // 当天日期作文件名
    char path[128];
    snprintf(path, sizeof(path), "%s%s.log", LOG_DIR, day);
    FILE *logfp = fopen(path, "a");
    if (!logfp) return;

    char ts[16];
    strftime(ts, sizeof(ts), "%H:%M:%S", &t);
    fprintf(logfp, "[%s] %s", ts, LOG_TAG);

    va_list args;
    va_start(args, fmt);
    vfprintf(logfp, fmt, args);
    va_end(args);
    fprintf(logfp, "\n");
    fclose(logfp);
}

void sig_handler(int s) {
    if (s == SIGINT || s == SIGTERM) running = 0;
    else if (s == SIGHUP) force_reload = 1;
}


// 读取 config.json，解析器见 common/json_config.h
using czero::parseJsonConfig;

// 把 Job 的日/月/周位图全置 1 表示不限，分/时由调度覆盖
static void job_init_full(Job* j){
    memset(j, 0, sizeof(*j));
    for(int i=0;i<24;i++) j->hour[i]=1;
    for(int i=1;i<=31;i++) j->day[i]=1;
    for(int i=1;i<=12;i++) j->month[i]=1;
    for(int i=0;i<7;i++) j->weekday[i]=1;
    j->day_spec=0; j->weekday_spec=0; j->pid=0;
}

// 每天 hh:mm 触发的 Job，用于定点任务和固定的 zero 任务
static void job_set_daily(Job* j, int hh, int mm){
    job_init_full(j);
    for(int i=0;i<24;i++) j->hour[i]=0;
    for(int i=0;i<60;i++) j->minute[i]=0;
    j->hour[hh]=1; j->minute[mm]=1;
}

// 把 every 和 at 换算成 Job 位图和门控秒，周期大于 1 天靠时间戳门控
// 返回 1 成功，0 无效或越界
static int schedule_to_job(const std::string& every, const std::string& at,
                           Job* j, long& gate_seconds){
    gate_seconds = 0;
    if(every.size()<3 || every[0]!='P') return 0;
    char unit = every[every.size()-1];
    if(every[1]=='T'){
        int n = atoi(every.substr(2, every.size()-3).c_str());
        if(n<=0) return 0;
        // 分钟上限 59，小时上限 23，越界无法表达返回 0
        if(unit=='M'){
            if(n>59) return 0;
            job_init_full(j);
            for(int m=0;m<60;m++) j->minute[m]=0;
            for(int m=0;m<60;m+=n) j->minute[m]=1;
            return 1;
        }
        if(unit=='H'){
            if(n>23) return 0;
            job_init_full(j);
            for(int h=0;h<24;h++) j->hour[h]=0;
            j->minute[0]=1;
            for(int h=0;h<24;h+=n) j->hour[h]=1;
            return 1;
        }
        return 0;
    }
    int n = atoi(every.substr(1, every.size()-2).c_str());
    if(n<=0) return 0;
    int hh=0, mm=0;
    size_t c = at.find(':');
    if(c!=std::string::npos){ hh=atoi(at.substr(0,c).c_str()); mm=atoi(at.substr(c+1).c_str()); }
    if(hh<0||hh>23) hh=0;
    if(mm<0||mm>59) mm=0;
    long days = (unit=='W') ? (long)n*7 : (unit=='D' ? n : 0);
    if(days<=0) return 0;
    job_set_daily(j, hh, mm);
    if(days>1) gate_seconds = days*86400L;
    return 1;
}

// 从 config.json 构建任务表和门控表
int load_config_jobs(void){
    auto cfg = parseJsonConfig(CONFIG_FILE);
    g_logEnabled = (cfg.count("general.log") && cfg["general.log"]=="true");  // 与其它组件一致
    if(cfg.empty()){ write_log("配置为空/解析失败，保留现有任务"); return njobs; }

    static Job nj[MAX_JOBS];
    static Gate ng_arr[MAX_GATES];
    int n=0, ng=0;

    auto add = [&](const char* everyKey, const char* atKey, const std::string& script, const char* desc){
        auto it = cfg.find(everyKey);
        if(it==cfg.end() || it->second.empty()) return;
        if(n>=MAX_JOBS) return;
        std::string at;
        if(atKey && cfg.count(atKey)) at = cfg[atKey];
        long gate=0;
        if(!schedule_to_job(it->second, at, &nj[n], gate)){ write_log("无效调度(%s): %s", desc, it->second.c_str()); return; }
        strncpy(nj[n].cmd, script.c_str(), sizeof(nj[n].cmd)-1);
        nj[n].cmd[sizeof(nj[n].cmd)-1]=0;
        n++;
        if(gate>0 && ng<MAX_GATES){
            strncpy(ng_arr[ng].cmd, script.c_str(), sizeof(ng_arr[ng].cmd)-1);
            ng_arr[ng].cmd[sizeof(ng_arr[ng].cmd)-1]=0;
            ng_arr[ng].min_sec=gate; ng++;
        }
    };

    add("app_clean.detect_schedule.every", NULL, MODDIR_S "/list/Tencent/check", "阈值检测");
    add("suppress.detect_schedule.every", NULL, MODDIR_S "/list/suppress/suppress", "阈值压制");
    add("app_clean.other.schedule.every", "app_clean.other.schedule.at", MODDIR_S "/list/customize", "应用清理");
    add("empty_folder.schedule.every", "empty_folder.schedule.at", MODDIR_S "/list/Emptyfolder/emptyfolder", "空文件夹清理");
    { auto sc = cfg.find("gc.script"); if(sc!=cfg.end() && !sc->second.empty()) add("gc.schedule.every", NULL, sc->second, "GC清理"); }
    // 每天 0 点重置今日统计，不在配置内
    if(n<MAX_JOBS){
        job_set_daily(&nj[n], 0, 0);
        strncpy(nj[n].cmd, MODDIR_S "/list/zero", sizeof(nj[n].cmd)-1);
        nj[n].cmd[sizeof(nj[n].cmd)-1]=0;
        n++;
    }

    if(n==0){ write_log("配置未产出有效任务，保留现有"); return njobs; }

    // 按 cmd 续接运行中子进程的 PID
    for(int i=0;i<n;i++){
        nj[i].pid=0;
        for(int j=0;j<njobs;j++)
            if(jobs[j].pid>0 && strcmp(jobs[j].cmd, nj[i].cmd)==0){ nj[i].pid=jobs[j].pid; break; }
    }
    for(int i=0;i<n;i++) jobs[i]=nj[i];
    njobs=n;
    for(int i=0;i<ng;i++) gates[i]=ng_arr[i];
    ngates=ng;
    write_log("已从配置加载 %d 个任务, %d 个门控", njobs, ngates);
    return njobs;
}

// 加载保存上次运行时间戳
void load_states(void) {
    nstates = 0;
    FILE *fp = fopen(STATE_FILE, "r");
    if (!fp) return;
    char line[MAX_LINE];
    while (fgets(line, sizeof(line), fp) && nstates < MAX_GATES) {
        char *eq = strrchr(line, '=');
        if (!eq) continue;
        *eq = 0;
        strncpy(states[nstates].cmd, line, sizeof(states[nstates].cmd) - 1);
        states[nstates].cmd[sizeof(states[nstates].cmd) - 1] = 0;
        states[nstates].last_run = (time_t)atol(eq + 1);
        nstates++;
    }
    fclose(fp);
}

void save_states(void) {
    FILE *fp = fopen(STATE_FILE, "w");
    if (!fp) return;
    for (int i = 0; i < nstates; i++)
        fprintf(fp, "%s=%ld\n", states[i].cmd, (long)states[i].last_run);
    fclose(fp);
}

// 命令是否受门控，返回最小间隔秒，0 表示无门控
long gate_for(const char *cmd) {
    for (int i = 0; i < ngates; i++)
        if (strcmp(cmd, gates[i].cmd) == 0) return gates[i].min_sec;
    return 0;
}

time_t get_last_run(const char *cmd) {
    for (int i = 0; i < nstates; i++)
        if (strcmp(states[i].cmd, cmd) == 0) return states[i].last_run;
    return 0;
}

void set_last_run(const char *cmd, time_t ts) {
    for (int i = 0; i < nstates; i++)
        if (strcmp(states[i].cmd, cmd) == 0) { states[i].last_run = ts; save_states(); return; }
    if (nstates < MAX_GATES) {
        strncpy(states[nstates].cmd, cmd, sizeof(states[nstates].cmd) - 1);
        states[nstates].cmd[sizeof(states[nstates].cmd) - 1] = 0;
        states[nstates].last_run = ts;
        nstates++;
        save_states();
    }
}

void reload_check(void) {
    if (force_reload) {
        force_reload = 0;
        write_log("收到 SIGHUP，强制重新加载配置");
        load_config_jobs();
        struct stat st;
        if (stat(CONFIG_FILE, &st) == 0) mtime = st.st_mtime;
        return;
    }
    struct stat st;
    if (stat(CONFIG_FILE, &st) == 0) {
        if (mtime == 0) {
            mtime = st.st_mtime;
        } else if (st.st_mtime != mtime) {
            write_log("配置文件已修改，重新加载");
            load_config_jobs();
            mtime = st.st_mtime;
        }
    }
}

// 返回 1 表示成功启动，上次仍在跑或 spawn 失败返回 0
int exec_job(const char *cmd, int idx) {
    if (jobs[idx].pid > 0 && kill(jobs[idx].pid, 0) == 0) {
        write_log("任务 %d 正在运行 (PID %d)，跳过", idx, jobs[idx].pid);
        return 0;
    }

    jobs[idx].pid = 0;
    
    pid_t pid;
    char *argv[] = {"/system/bin/sh", "-c", (char*)cmd, NULL};
    
    posix_spawn_file_actions_t fa;
    posix_spawn_file_actions_init(&fa);
    
    int null_fd = open("/dev/null", O_WRONLY);
    if (null_fd >= 0) {
        posix_spawn_file_actions_adddup2(&fa, null_fd, STDOUT_FILENO);
        posix_spawn_file_actions_adddup2(&fa, null_fd, STDERR_FILENO);
        posix_spawn_file_actions_addclose(&fa, null_fd);
    }
    
    posix_spawnattr_t attr;
    posix_spawnattr_init(&attr);
    
    int ret = posix_spawn(&pid, "/system/bin/sh", &fa, &attr, argv, environ);
    
    posix_spawnattr_destroy(&attr);
    posix_spawn_file_actions_destroy(&fa);
    if (null_fd >= 0) close(null_fd);
    
    if (ret == 0) {
        jobs[idx].pid = pid;
        write_log("执行任务 %d (PID %d): %s", idx, pid, cmd);
        return 1;
    }
    write_log("任务 %d spawn 失败 (errno %d): %s", idx, ret, cmd);
    return 0;
}

int should_run(Job *job, struct tm *tm) {
    if (!job->minute[tm->tm_min]) return 0;
    if (!job->hour[tm->tm_hour]) return 0;
    if (!job->month[tm->tm_mon + 1]) return 0;
    
    if (job->day_spec && job->weekday_spec) {
        if (!job->day[tm->tm_mday] && !job->weekday[tm->tm_wday]) return 0;
    } else if (job->day_spec) {
        if (!job->day[tm->tm_mday]) return 0;
    } else if (job->weekday_spec) {
        if (!job->weekday[tm->tm_wday]) return 0;
    }
    
    return 1;
}

// 上次处理到的分钟点，用于休眠漏触发后补跑
static time_t last_tick = 0;
#define MAX_CATCHUP_MIN 1440

void check_jobs(void) {
    // 先收尸再跑任务，避免僵尸进程被误判为仍在运行
    int status;
    pid_t p;
    while ((p = waitpid(-1, &status, WNOHANG)) > 0) {
        for (int i = 0; i < njobs; i++) {
            if (jobs[i].pid == p) {
                jobs[i].pid = 0;
                break;
            }
        }
    }

    reload_check();

    time_t now = time(NULL);
    time_t cur_min = now - (now % 60);
    if (last_tick == 0) last_tick = cur_min - 60;  // 首次只处理当前分钟

    time_t start = last_tick + 60;
    if (cur_min - start > (time_t)MAX_CATCHUP_MIN * 60) {
        write_log("检测到 %ld 分钟缺口，仅补跑最近 %d 分钟",
                  (long)((cur_min - start) / 60), MAX_CATCHUP_MIN);
        start = cur_min - (time_t)MAX_CATCHUP_MIN * 60;
    }

    // 跨所有遗漏分钟点求并集，每个任务在缺口内命中一次就跑一次
    char due[MAX_JOBS];
    memset(due, 0, sizeof(due));
    for (time_t t = start; t <= cur_min; t += 60) {
        struct tm tm;
        localtime_r(&t, &tm);
        for (int i = 0; i < njobs; i++)
            if (!due[i] && should_run(&jobs[i], &tm)) due[i] = 1;
    }
    last_tick = cur_min;

    for (int i = 0; i < njobs; i++) {
        if (!due[i]) continue;
        long g = gate_for(jobs[i].cmd);
        if (g > 0) {
            time_t last = get_last_run(jobs[i].cmd);
            if (last != 0 && (now - last) < g) {
                write_log("门控跳过 %s (距上次 %ld 秒 < %ld)", jobs[i].cmd, (long)(now - last), g);
                continue;
            }
            // 仅在真正启动后才刷新门控时间戳
            if (exec_job(jobs[i].cmd, i))
                set_last_run(jobs[i].cmd, now);
        } else {
            exec_job(jobs[i].cmd, i);
        }
    }
}

int create_timer(void) {
    int fd = timerfd_create(CLOCK_BOOTTIME_ALARM, TFD_NONBLOCK | TFD_CLOEXEC);
    if (fd < 0) {
        fd = timerfd_create(CLOCK_BOOTTIME, TFD_NONBLOCK | TFD_CLOEXEC);
        if (fd < 0) {
            fd = timerfd_create(CLOCK_MONOTONIC, TFD_NONBLOCK | TFD_CLOEXEC);
            if (fd < 0) return -1;
        }
    }
    
    struct tm tm;
    get_local_time(&tm);
    int sec = 60 - tm.tm_sec;
    
    struct itimerspec spec = {{60, 0}, {sec, 0}};
    if (timerfd_settime(fd, 0, &spec, NULL) < 0) {
        close(fd);
        return -1;
    }
    
    return fd;
}

int get_lock(void) {
    lockfd = open(LOCK_FILE, O_CREAT | O_RDWR, 0644);
    if (lockfd < 0) return -1;
    
    if (flock(lockfd, LOCK_EX | LOCK_NB) < 0) {
        close(lockfd);
        return -1;
    }

    if (ftruncate(lockfd, 0) == 0) {
        char buf[32];
        snprintf(buf, sizeof(buf), "%d\n", getpid());
        if (write(lockfd, buf, strlen(buf)) < 0) { }
    }
    return 0;
}

void release_lock(void) {
    if (lockfd >= 0) {
        flock(lockfd, LOCK_UN);
        close(lockfd);
        unlink(LOCK_FILE);
    }
}

int main(int argc, char *argv[]) {
    int fg = (argc > 1 && strcmp(argv[1], "-f") == 0);
    
    mkdir("/data/adb/modules/CZero/log", 0777);
    mkdir("/data/adb/modules/CZero/cron", 0777);
    
    FILE *fp = fopen("/proc/self/oom_score_adj", "w");
    if (fp) {
        fprintf(fp, "650"); 
        fclose(fp);
    }

    
    if (get_lock() < 0) return 1;
    
    if (!fg && daemon(0, 0) < 0) {
        release_lock();
        return 1;
    }
    
    // 先解析一次配置以确定日志开关
    { auto c = parseJsonConfig(CONFIG_FILE); g_logEnabled = (c.count("general.log") && c["general.log"]=="true"); }
    write_log("Timer 启动 ！(PID %d)", getpid());

    signal(SIGINT, sig_handler);
    signal(SIGTERM, sig_handler);
    signal(SIGHUP, sig_handler);
    signal(SIGPIPE, SIG_IGN);
    
    load_config_jobs();
    // 记录 config.json 当前 mtime 作为基线，之后仅在其变化时重载
    { struct stat st; if (stat(CONFIG_FILE, &st) == 0) mtime = st.st_mtime; }
    load_states();
    
    int tfd = create_timer();
    if (tfd < 0) {
        write_log("创建定时器失败");
        release_lock();
        return 1;
    }
    
    int efd = epoll_create1(EPOLL_CLOEXEC);
    if (efd < 0) {
        close(tfd);
        release_lock();
        return 1;
    }
    
    struct epoll_event ev = {.events = EPOLLIN, .data.fd = tfd};
    if (epoll_ctl(efd, EPOLL_CTL_ADD, tfd, &ev) < 0) {
        close(efd);
        close(tfd);
        release_lock();
        return 1;
    }
    
    write_log("守护进程运行中...");
    
    struct epoll_event events[10];
    uint64_t exp;
    
    while (running) {
        int n = epoll_wait(efd, events, 10, -1);
        if (n < 0) {
            if (errno == EINTR) continue;
            break;
        }
        
        for (int i = 0; i < n; i++) {
            if (events[i].data.fd == tfd) {
                if (read(tfd, &exp, sizeof(exp)) == sizeof(exp)) {
                    check_jobs();
                }
            }
        }
    }
    
    write_log("守护进程停止");
    close(efd);
    close(tfd);
    release_lock();
    
    return 0;
}