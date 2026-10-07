# 下载

<script setup>
import rel from './.vitepress/release.json'
const MIRROR = 'https://dl.czeropage.top/'
const GH = 'https://github.com/Xocio/CZero/releases'
const files = [
  { label: 'CZero 模块', fit: 'Root 版', a: rel.module },
  { label: 'CZeroX 应用', fit: 'Root 版', a: rel.app },
  { label: 'CZeroX 应用', fit: 'Shizuku 版', a: rel.shizuku },
].map((f) => ({ ...f, file: f.a?.file || '', sha: f.a?.sha256 || '', url: f.a?.file ? MIRROR + f.a.file : '' }))
</script>

::: tip 国内用户请走直链
GitHub 在部分网络环境下无法访问或速度极慢，本站提供**国内直链**，无需代理即可下载。两个渠道分发的是同一份文件，可用下方 SHA-256 自行校验。
:::

<DownloadEditions inline />

## 全部文件

<table class="dl-table">
  <thead>
    <tr><th>组件</th><th>适用</th><th>文件</th><th>下载</th></tr>
  </thead>
  <tbody>
    <tr v-for="f in files" :key="f.label + f.fit">
      <td><strong>{{ f.label }}</strong></td>
      <td>{{ f.fit }}</td>
      <td><code v-if="f.file">{{ f.file }}</code><span v-else class="dl-na">暂未提供</span></td>
      <td><template v-if="f.url"><a :href="f.url">国内直链</a> · <a :href="GH">GitHub</a></template><span v-else class="dl-na">—</span></td>
    </tr>
  </tbody>
</table>

::: details 文件校验（SHA-256）

<div class="dl-sums">
  <template v-for="f in files" :key="f.file">
    <div v-if="f.file" class="dl-sum"><span>{{ f.file }}</span><code>{{ f.sha }}</code></div>
  </template>
</div>

请勿使用来源不明的二次打包版本，模块与应用均以系统权限运行。校验方式：

- **手机上** —— 用 MT 管理器等文件管理器查看文件的 SHA-256。
- **电脑上** —— Windows 执行 `certutil -hashfile 文件名 SHA256`，Linux / macOS 执行 `sha256sum 文件名`。

:::

## 环境要求

两个版本的完整区别见[版本选择](/guide/editions)。

**Root 版**

- Android 9+（API 28），`arm64-v8a`
- Root 方案：Magisk、KernelSU 或 APatch
- F2FS `/data` 分区（仅 GC 功能需要，其余功能不受影响）
- LSPosed（仅精细化压制需要，其余功能不受影响）

**Shizuku 版**

- `arm64-v8a` 设备
- 已安装并启动 [Shizuku](https://shizuku.rikka.app/download/)，系统版本以其要求为准

## 安装

**Root 版**

1. 在 Magisk / KernelSU / APatch 中刷入模块 zip，按音量键提示选择语言与是否继承旧配置。
2. 重启设备。CZeroX 已内置在模块中，刷入时自动安装；若未能自动安装，可在上表单独下载 Root 版应用。

**Shizuku 版**

1. 安装并启动 Shizuku。
2. 安装 CZeroX（Shizuku 版），打开后按提示向 Shizuku 申请授权，应用会自动完成清理程序与规则的部署。

完整步骤与常见问题见 [安装与上手](/guide/getting-started)。

::: warning 两版不可同时安装
两版使用同一个应用标识，安装其中一版会直接替换另一版。Shizuku 版用户请勿刷入 CZero 模块，模块会自动安装 Root 版应用并将其替换。
:::

::: tip 更新时注意
- **Root 版** —— 刷入新版模块时选择**继承配置**，会保留黑白名单与自定义路径列表，新版新增的配置项会自动补全到 `config.json` 中。
- **Shizuku 版** —— 覆盖安装新版应用即可，打开应用后会部署新版清理程序与规则。
:::

## 下载遇到问题？

- **直链也下不动** —— 换用移动数据或 Wi-Fi 再试一次；仍不行可改用 GitHub 渠道，或在应用内检查更新直接下载。
- **浏览器提示文件有风险** —— 模块 zip 与 apk 属于正常提示，核对 SHA-256 一致即可放心使用。
- **不确定下载哪个文件** —— 文件名中带 `Root` 的为 Root 版应用，带 `Shizuku` 的为 Shizuku 版应用；Root 版用户通常只需下载模块。
- **装不上 / 刷入失败** —— 先确认设备满足上方环境要求，再查阅 [常见问题](/guide/faq)。
