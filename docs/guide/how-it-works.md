# 工作原理

本页深入说明 CZero 从开机到执行清理的完整链路，帮助你理解它"为什么这样设计"。

## 总览

CZero 没有常驻的 Java 服务，而是把职责拆成三层：**安装期**准备文件，**开机期**拉起进程，**运行期**由一个 C++ 守护进程按配置调度所有任务。

```
安装 (customize.sh)  ──►  开机 (service.sh → service)  ──►  运行 (timer_daemon)
   选语言/继承配置          初始化/改权限/拉起守护进程        读 config.json 调度清理
```

## 安装期：customize.sh

刷入时运行一次，主要做三件事：

1. **语言选择** —— 通过 `getevent` 读音量键：音量上 = English，音量下 = 中文；约 60 秒无按键则默认中文。之后所有提示都用所选语言。
2. **继承旧配置**（可选）—— 若选择继承且检测到旧安装，仅迁移以下用户数据，其余重置为新版默认：
   - `list/Emptyfolder/directories.prop`、`emptyfolder_white.prop`
   - `list/clean_whitelist.prop`、`list/clean_paths.prop`
   - `basis/` 统计目录
   - 注意：`config.json` **不继承**，始终使用新版默认值。
3. **打开项目主页** —— 安装完成后拉起浏览器指向 GitHub 仓库。

## 开机期：service.sh → service

每次开机运行：

1. `service.sh` 先修正权限（目录 755，`.prop`/`.json`/`.log` 等 644），再启动 `service` 二进制，然后 `pkill` 旧的 `timer_daemon` 并用 `setsid` 重新拉起。
2. `service`（C++）负责：更新 `module.prop` 的 `description` 显示当前状态、必要时创建 `basis/basis.prop` 统计文件、修正模块内文件权限。它**不生成任何调度文件** —— 调度完全由守护进程直接读 `config.json` 得出。

## 运行期：cron/timer_daemon

这是 CZero 的核心 —— 一个用 `timerfd` + `epoll` 实现的极简 cron 引擎：

- **直接读 `config.json`** —— 内置 JSON 解析与 ISO8601→cron 转换，在内存里构建任务表与门控表，不落盘成中间文件。
- **热重载** —— 监视 `config.json` 的 mtime，变更后在下一分钟自动重建任务集；若配置为空或无法解析，则保留上一份有效任务集，服务不中断。
- **固定日任务** —— 额外注册一个每天 00:00 的 `list/zero` 任务（重置统计、清理非今日日志），该任务不在 `config.json` 中。
- **逐分钟触发** —— 按手机本地时间（`localtime_r`）每分钟检查一次；用 `posix_spawn` 运行脚本；若上一轮还在执行则跳过本轮（防重入）。
- **多天任务门控** —— 周期 ≥ 1 天的任务，其上次运行时间戳持久化在 `cron/state`，未满周期则跳过。
- **单实例** —— 通过 `cron/timer.lock` 防止重复启动。
- **日志** —— 与其他组件共写每日日志（标签 `[定时]`），同样受 `general.log` 开关控制。

## 一次清理是如何发生的

以微信缓存清理为例：

1. 守护进程到点触发 `list/Tencent/check`。
2. `check` 识别前台应用：读 `/dev/cpuset/top-app` 与 `dumpsys activity` 的 `ResumedActivity`。
3. 判断保护条件：前台是微信本身、或前台是游戏 → 本轮跳过；否则继续。
4. 调用对应清理器（`tencentmm` 等 Go 二进制）执行清理，受 `enhanced` 与时序屏障约束。
5. 更新 `basis.prop` 统计，按需写日志，并可向 CZeroX 广播进度 / 结果。

## 为什么这样设计

- **无常驻服务** —— 只在需要时被守护进程唤起，平时不占内存与 CPU。
- **单一配置源 + 热重载** —— 配置与行为一一对应，改完即生效，且坏配置不会拖垮服务。
- **原生前端解耦** —— CZeroX 只读写 `config.json`，与模块彻底解耦，模块不装应用也能独立运行。
