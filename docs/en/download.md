# Download

<script setup>
import rel from '../.vitepress/release.json'
const MIRROR = 'https://dl.czeropage.top/'
const moduleUrl = MIRROR + rel.module.file
const appUrl = MIRROR + rel.app.file
</script>

::: tip Mirror available
GitHub is slow or unreachable on some networks. This site serves the same files from its own mirror — both channels distribute identical builds, verifiable with the SHA-256 checksums below.
:::

## Latest release

Current version <span class="dl-ver">{{ rel.version }}</span>, released {{ rel.date }}.

The module and the app are meant to be used together: the module does the actual cleaning and scheduling, while **CZeroX** lets you check status and change settings. Install both on a fresh setup.

| Component | File | Download |
|---|---|---|
| **CZero** module | <code>{{ rel.module.file }}</code> | <a :href="moduleUrl">Mirror</a> · <a href="https://github.com/Xocio/CZero/releases">GitHub</a> |
| **CZeroX** app | <code>{{ rel.app.file }}</code> | <a :href="appUrl">Mirror</a> · <a href="https://github.com/Xocio/CZero/releases">GitHub</a> |

::: details Checksums (SHA-256)

<div class="dl-sums">
  <div class="dl-sum"><span>{{ rel.module.file }}</span><code>{{ rel.module.sha256 }}</code></div>
  <div class="dl-sum"><span>{{ rel.app.file }}</span><code>{{ rel.app.sha256 }}</code></div>
</div>

The module runs with root privileges — never install a repackaged build from an untrusted source. To verify:

- **On device** — check the file's SHA-256 in a file manager such as MT Manager.
- **On desktop** — run `certutil -hashfile FILENAME SHA256` on Windows, or `sha256sum FILENAME` on Linux / macOS.

:::

## Requirements

- Android 9+ (API 28), `arm64-v8a`
- Root solution: Magisk, KernelSU or APatch
- F2FS `/data` partition (only needed for GC; other features work regardless)

## Installation

1. Flash the module zip in Magisk / KernelSU / APatch, following the volume-key prompts for language and config migration.
2. Reboot the device.
3. Install the CZeroX app.

See [Install & Setup](/en/guide/getting-started) for full steps and troubleshooting.

::: warning When updating
Choosing to **migrate config** while flashing keeps your allow/block lists and custom paths, and any options added in the new version are merged into your `config.json` automatically.
:::

## Download issues

- **Mirror is slow too** — try switching between mobile data and Wi-Fi; otherwise use the GitHub channel, or let the app download the update directly.
- **Browser warns about the file** — expected for module zips and apks; matching SHA-256 means the file is intact.
- **Fails to install or flash** — confirm your device meets the requirements above, then check the [FAQ](/en/guide/faq).
