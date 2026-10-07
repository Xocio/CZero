<p align="center">
  <img src="assets/logo.png" width="120" alt="CZero logo">
</p>

<h1 align="center">CZero</h1>

<p align="center">力争做 Android 最好的缓存垃圾清理方案。</p>

<p align="center">
  <a href="https://github.com/Xocio/CZero/releases"><img src="https://img.shields.io/github/v/release/Xocio/CZero?label=Release&color=orange" alt="Release"></a>
  <a href="https://czeropage.top/"><img src="https://img.shields.io/badge/%E6%96%87%E6%A1%A3-Docs-blue" alt="Docs"></a>
  <img src="https://img.shields.io/badge/ROOT-Magisk%20%7C%20KernelSU%20%7C%20APatch-red" alt="Root">
  <img src="https://img.shields.io/badge/%E5%85%8D%20ROOT-Shizuku-2E7D32" alt="Shizuku">
  <a href="https://t.me/CZeroRelease"><img src="https://img.shields.io/badge/Telegram-频道-26A5E4?logo=telegram&logoColor=white" alt="Telegram Channel"></a>
</p>

<p align="center"><b>简体中文</b> · <a href="README_en.md">English</a> · <a href="README_ru.md">Русский</a></p>

---

CZero 是一套 Android 清理方案，为常见的高频应用提供缓存清理，并涵盖后台压制、精细化压制、空文件夹清理、文件归类、F2FS 垃圾回收与 fstrim 等功能。

没有常驻服务，所有任务由一个极轻量的原生调度进程按 `config.json` 触发，配置修改即时生效。日常操作通过原生配套应用 **CZeroX** 完成。

## 两个版本

| | **Root 版** | **Shizuku 版** |
|---|---|---|
| 授权方式 | Magisk / KernelSU / APatch | Shizuku |
| 组成 | CZero 模块 + CZeroX（模块内置，刷入后自动安装） | 仅 CZero，清理程序随应用自带 |
| 功能范围 | 全部功能 | 不含精细化压制与存储面板；F2FS 回收由系统空闲维护完成 |

两版使用同一个应用标识，不能同时安装。详细对比见 [版本选择](https://czeropage.top/guide/editions)。

## 社区

- **发布频道**：[Release](https://t.me/CZeroRelease) —— 版本更新与公告。
- **交流群聊**：[Organize](https://t.me/+lwNKCHw_NktjODRh) —— 使用反馈与讨论。

## 文档

> **官方文档  [DOCS](https://czeropage.top/)**
>
> 使用前强烈建议先阅读文档，遇到问题请优先查看 [常见问题](https://czeropage.top/guide/faq)。

## 功能

- **定向缓存清理** —— 微信 / QQ / 抖音各自独立的清理器，覆盖应用双开；清理前检测前台与游戏状态，同一应用两次清理之间保持最小间隔。
- **后台压制** —— 周期性结束目标应用的后台子进程，主进程与消息推送不受影响。
- **精细化压制**（Root 版，BETA） —— 挂起后台应用而不结束进程，运行状态完整保留、切回即时恢复，并提供生效检测。
- **文件归类** —— 按类型将下载目录等位置的文件整理进分类目录，支持归类历史与完整还原。
- **F2FS 垃圾回收与 fstrim** —— 仅在设备真正空闲时执行，维持存储空间与长期写入性能。
- **自定义规则与规则源** —— 自行声明清理路径与白名单，或订阅第三方规则源，每天自动更新。
- **回收站** —— 清理先移入回收站，默认保留 7 天，可原样恢复。
- **Zero 智能助手** —— 以对话方式扫描存储、预估清理效果、管理规则与压制，任何更改都需确认后执行。
- **配置热重载** —— 修改即时生效；配置损坏时保留上一份有效任务，不会中断。

## 下载与安装

前往 [下载页](https://czeropage.top/download-home)（国内直链）或 [Releases](https://github.com/Xocio/CZero/releases) 获取最新版本。

| 文件 | 说明 |
|---|---|
| `CZero_<版本>.zip` | Root 版模块，已内置 Root 版 CZeroX |
| `CZeroX_Root_<版本>.apk` | Root 版应用，仅在刷入后未能自动安装时使用 |
| `CZero_Shizuku_<版本>.apk` | Shizuku 版应用 |

**Root 版**

1. 用 Magisk / KernelSU / APatch 刷入模块 zip，按音量键提示选择语言、是否继承旧配置。
2. 重启，CZeroX 随模块自动安装。

**Shizuku 版**

1. 安装并启动 [Shizuku](https://shizuku.rikka.app/download/)。
2. 安装 CZeroX（Shizuku 版），打开后按提示授权，应用会自动完成部署。

## CZeroX

<table>
<tr>
<td valign="top" width="50%">

配套的原生 Jetpack Compose 应用，界面采用 [Miuix](https://compose-miuix-ui.github.io/miuix/) 风格，支持简体中文、English 与 Русский。

</td>
<td align="center" width="50%">
<img src="assets/webx.png" width="220" alt="CZeroX 主页">
</td>
</tr>
</table>

## Star History

<a href="https://www.star-history.com/?repos=Xocio%2FCZero&type=timeline&logscale=&legend=top-left">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/chart?repos=Xocio/CZero&type=timeline&theme=dark&logscale&legend=top-left&sealed_token=twVK7kU7SjickXCW34YQxO2BJE8Ll27rIB3db1HiNE9oyq1tMAXVJy3TiSVIlrdDuAeF0VGVZEdJTbr2bIBoyyvYERJyDzdmRNbeOOwKSMJZRyid1w3R1pxSIclT5LPro3oFtNGwvcdokYqwmWLAIVDeIo_axyrSqJsR1o8BY-_KOHqAIEWhs6lAn4fa" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/chart?repos=Xocio/CZero&type=timeline&logscale&legend=top-left&sealed_token=twVK7kU7SjickXCW34YQxO2BJE8Ll27rIB3db1HiNE9oyq1tMAXVJy3TiSVIlrdDuAeF0VGVZEdJTbr2bIBoyyvYERJyDzdmRNbeOOwKSMJZRyid1w3R1pxSIclT5LPro3oFtNGwvcdokYqwmWLAIVDeIo_axyrSqJsR1o8BY-_KOHqAIEWhs6lAn4fa" />
   <img alt="Star History Chart" src="https://api.star-history.com/chart?repos=Xocio/CZero&type=timeline&logscale&legend=top-left&sealed_token=twVK7kU7SjickXCW34YQxO2BJE8Ll27rIB3db1HiNE9oyq1tMAXVJy3TiSVIlrdDuAeF0VGVZEdJTbr2bIBoyyvYERJyDzdmRNbeOOwKSMJZRyid1w3R1pxSIclT5LPro3oFtNGwvcdokYqwmWLAIVDeIo_axyrSqJsR1o8BY-_KOHqAIEWhs6lAn4fa" />
 </picture>
</a>

## 许可证

[GNU General Public License v3.0](LICENSE)
