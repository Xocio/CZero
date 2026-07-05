# 安装与上手

## 环境要求

- Android 9+（API 28），`arm64-v8a`
- Root 方案：Magisk、KernelSU 或 APatch
- F2FS `/data` 分区（仅 GC 功能需要，其余功能不受影响）

## 安装步骤

1. 从 [Releases](https://github.com/Xocio/CZero/releases) 下载最新模块 zip。
2. 在 Magisk / KernelSU / APatch 中刷入该 zip。
3. 刷入过程中按音量键提示操作：

   | 提示 | 操作 |
   |---|---|
   | 选择语言 | 音量上 = English / 音量下 = 中文 |
   | 是否继承旧配置 | 音量键选择 Y / N |

   > 继承时只沿用黑白名单与自定义路径列表；`config.json` 始终重置为新版默认值。

4. 重启设备。
5. 安装 **CZeroX** 应用，用于查看状态与调整配置。

::: tip 首次安装
不确定选什么就全部使用默认值即可 —— 默认配置已经开启缓存清理、后台压制、GC 与空文件夹清理，开箱即用。
:::

## 验证是否生效

重启后可通过以下任一方式确认模块在运行：

- 在 Root 管理器中查看 CZero 模块的描述，应显示为已生效状态。
- 打开 **CZeroX**，主页会显示守护进程状态（PID / 内存占用）与清理统计。

## 更新模块

下载新版 zip 后直接刷入覆盖即可。刷入过程中若选择**继承配置**，会保留你的黑白名单与自定义路径列表；但 `config.json` 始终重置为新版默认值 —— 更新后如有需要请重新调整或用 CZeroX 重设。

## 卸载

在 Root 管理器中移除 CZero 模块并重启即可。模块目录 `/data/adb/modules/CZero/` 会被清除，含其中的统计与日志。

## 安装遇到问题？

- **音量键没反应 / 未检测到按键** —— 部分设备 `getevent` 响应较慢，安装器约 60 秒无按键会默认中文并继续，不会卡住。
- **刷入后不生效** —— 确认已重启；再到 CZeroX 主页查看守护进程状态，必要时一键重启进程。
- **GC 不工作** —— 确认 `/data` 为 F2FS 分区，且 `gc.enabled` 为 `true`。非 F2FS 设备请关闭 GC。

更多问题见 [常见问题](/guide/faq)。

## 下一步

- [工作原理](/guide/how-it-works) —— 从开机到清理的完整链路。
- [功能详解](/guide/features) —— 了解每个功能具体做了什么。
- [配置参考](/guide/configuration) —— 逐字段说明 `config.json`。
- [调度模型](/guide/schedule) —— 理解 `every` / `at` 的写法与规则。
