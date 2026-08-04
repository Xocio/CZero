# 下载

<script setup>
import rel from './.vitepress/release.json'
const MIRROR = 'https://dl.czeropage.top/'
const moduleUrl = MIRROR + rel.module.file
const appUrl = MIRROR + rel.app.file
</script>

::: tip 国内用户请走直链
GitHub 在部分网络环境下无法访问或速度极慢，本站提供**国内直链**，无需代理即可下载。两个渠道分发的是同一份文件，可用下方 SHA-256 自行校验。
:::

## 最新版本

当前版本 <span class="dl-ver">{{ rel.version }}</span>，发布于 {{ rel.date }}。

模块与应用需要配套使用：模块负责实际的清理与调度，**CZeroX** 用于查看状态和修改配置。首次安装请两者都装。

| 组件 | 文件 | 下载 |
|---|---|---|
| **CZero** 模块 | <code>{{ rel.module.file }}</code> | <a :href="moduleUrl">国内直链</a> · <a href="https://github.com/Xocio/CZero/releases">GitHub</a> |
| **CZeroX** 应用 | <code>{{ rel.app.file }}</code> | <a :href="appUrl">国内直链</a> · <a href="https://github.com/Xocio/CZero/releases">GitHub</a> |

::: details 文件校验（SHA-256）

<div class="dl-sums">
  <div class="dl-sum"><span>{{ rel.module.file }}</span><code>{{ rel.module.sha256 }}</code></div>
  <div class="dl-sum"><span>{{ rel.app.file }}</span><code>{{ rel.app.sha256 }}</code></div>
</div>

模块以 root 权限运行，请勿使用来源不明的二次打包版本。校验方式：

- **手机上** —— 用 MT 管理器等文件管理器查看文件的 SHA-256。
- **电脑上** —— Windows 执行 `certutil -hashfile 文件名 SHA256`，Linux / macOS 执行 `sha256sum 文件名`。

:::

## 环境要求

- Android 9+（API 28），`arm64-v8a`
- Root 方案：Magisk、KernelSU 或 APatch
- F2FS `/data` 分区（仅 GC 功能需要，其余功能不受影响）

## 安装

1. 在 Magisk / KernelSU / APatch 中刷入模块 zip，按音量键提示选择语言与是否继承旧配置。
2. 重启设备。
3. 安装 CZeroX 应用。

完整步骤与常见问题见 [安装与上手](/guide/getting-started)。

::: warning 更新时注意
刷入新版时若选择**继承配置**，会保留黑白名单与自定义路径列表，新版新增的配置项会自动补全到你的 `config.json` 中。
:::

## 下载遇到问题？

- **直链也下不动** —— 换用移动数据或 Wi-Fi 再试一次；仍不行可改用 GitHub 渠道，或在应用内检查更新直接下载。
- **浏览器提示文件有风险** —— 模块 zip 与 apk 属于正常提示，核对 SHA-256 一致即可放心使用。
- **装不上 / 刷入失败** —— 先确认设备满足上方环境要求，再查阅 [常见问题](/guide/faq)。
