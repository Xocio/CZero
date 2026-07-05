# 什么是 CZero

CZero 是一个 Android Root 清理模块，为常见的高频应用提供缓存清理，并涵盖后台压制、空文件夹清理与 F2FS 垃圾回收。

模块本身没有常驻服务，由一个轻量 C++ 守护进程按 `config.json` 调度所有任务，配置修改即时生效。日常操作通过原生配套应用 **CZeroX** 完成。

## 它解决什么问题

长期使用后，应用缓存与后台常驻会持续占用存储和内存，手机逐渐变卡、可用空间变少。CZero 在后台按计划自动完成清理与压制，让设备保持干净、流畅，无需手动干预。

## 设计理念

- **无常驻服务** —— 不驻留任何 Java 进程，调度全部交给一个基于 `timerfd` + `epoll` 的微型 C++ 守护进程，开销趋近于零。
- **单一配置源** —— 所有行为集中在一份 `config.json`，守护进程与各清理组件直接读取，无中间派生文件。
- **即时生效** —— 守护进程监视 `config.json` 的变更并热重载；配置损坏时保留上一份有效任务，绝不中断。
- **原生前端** —— 日常操作通过原生应用 CZeroX 完成，不依赖 WebUI。

## 三层架构

CZero 由三个协同工作的层组成：

| 层 | 组件 | 职责 |
|---|---|---|
| 安装 | `customize.sh` | 音量键选择语言、询问是否继承旧配置 |
| 开机 | `service.sh` → `service`（C++） | 标记模块生效、初始化统计文件、修正权限、拉起守护进程 |
| 运行时 | `cron/timer_daemon`（C++） | 解析 `config.json`、调度并派生全部清理任务、变更即热重载 |

## 运行时文件

所有运行时文件位于 `/data/adb/modules/CZero/`：

```text
config.json            # 唯一配置源（守护进程与各清理组件直接读取）
basis/basis.prop       # 清理统计
cron/state             # ≥1 天任务的上次执行时间
list/Tencent/…         # 高频应用检测与清理脚本
list/suppress/…        # 后台压制
list/GCclean/…         # F2FS GC 脚本
list/Emptyfolder/…     # 空文件夹清扫
log/<YYYY-MM-DD>.log   # 单一每日日志（仅保留当天）
```

## 环境要求

- Android 9+（API 28），`arm64-v8a`
- Root 方案：Magisk、KernelSU 或 APatch
- F2FS `/data` 分区（仅 GC 功能需要，其余功能不受影响）

准备好后，前往 [安装与上手](/guide/getting-started)。
