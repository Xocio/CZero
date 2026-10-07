# Download

<script setup>
import rel from '../.vitepress/release.json'
const MIRROR = 'https://dl.czeropage.top/'
const GH = 'https://github.com/Xocio/CZero/releases'
const files = [
  { label: 'CZero module', fit: 'Root edition', a: rel.module },
  { label: 'CZeroX app', fit: 'Root edition', a: rel.app },
  { label: 'CZeroX app', fit: 'Shizuku edition', a: rel.shizuku },
].map((f) => ({ ...f, file: f.a?.file || '', sha: f.a?.sha256 || '', url: f.a?.file ? MIRROR + f.a.file : '' }))
</script>

::: tip Mirror available
GitHub is slow or unreachable on some networks. This site serves the same files from its own mirror — both channels distribute identical builds, verifiable with the SHA-256 checksums below.
:::

<DownloadEditions inline />

## All files

<table class="dl-table">
  <thead>
    <tr><th>Component</th><th>For</th><th>File</th><th>Download</th></tr>
  </thead>
  <tbody>
    <tr v-for="f in files" :key="f.label + f.fit">
      <td><strong>{{ f.label }}</strong></td>
      <td>{{ f.fit }}</td>
      <td><code v-if="f.file">{{ f.file }}</code><span v-else class="dl-na">Not available yet</span></td>
      <td><template v-if="f.url"><a :href="f.url">Mirror</a> · <a :href="GH">GitHub</a></template><span v-else class="dl-na">—</span></td>
    </tr>
  </tbody>
</table>

::: details Checksums (SHA-256)

<div class="dl-sums">
  <template v-for="f in files" :key="f.file">
    <div v-if="f.file" class="dl-sum"><span>{{ f.file }}</span><code>{{ f.sha }}</code></div>
  </template>
</div>

Both the module and the app run with system privileges — never install a repackaged build from an untrusted source. To verify:

- **On device** — check the file's SHA-256 in a file manager such as MT Manager.
- **On desktop** — run `certutil -hashfile FILENAME SHA256` on Windows, or `sha256sum FILENAME` on Linux / macOS.

:::

## Requirements

See [Choosing an Edition](/en/guide/editions) for the full differences between the two editions.

**Root edition**

- Android 9+ (API 28), `arm64-v8a`
- Root solution: Magisk, KernelSU or APatch
- F2FS `/data` partition (only needed for GC; other features work regardless)
- LSPosed (only needed for fine-grained suppression; other features work regardless)

**Shizuku edition**

- An `arm64-v8a` device
- [Shizuku](https://shizuku.rikka.app/download/) installed and running; system version as required by it

## Installation

**Root edition**

1. Flash the module zip in Magisk / KernelSU / APatch, following the volume-key prompts for language and config migration.
2. Reboot the device. CZeroX ships inside the module and is installed while flashing; if that fails, download the Root edition app from the table above.

**Shizuku edition**

1. Install and start Shizuku.
2. Install CZeroX (Shizuku edition), open it and request access from Shizuku when prompted; the app deploys the cleaning programs and rules by itself.

See [Install & Setup](/en/guide/getting-started) for full steps and troubleshooting.

::: warning The editions cannot be installed side by side
Both editions share the same app ID, so installing one replaces the other. Shizuku-edition users should not flash the CZero module: it installs the Root edition app automatically, replacing theirs.
:::

::: tip When updating
- **Root edition** — choosing to **migrate config** while flashing keeps your allow/block lists and custom paths, and options added in the new version are merged into `config.json` automatically.
- **Shizuku edition** — install the new app over the old one; opening it deploys the new cleaning programs and rules.
:::

## Download issues

- **Mirror is slow too** — try switching between mobile data and Wi-Fi; otherwise use the GitHub channel, or let the app download the update directly.
- **Browser warns about the file** — expected for module zips and apks; matching SHA-256 means the file is intact.
- **Not sure which file to download** — files with `Root` in the name are the Root edition app, files with `Shizuku` the Shizuku edition app; Root-edition users usually only need the module.
- **Fails to install or flash** — confirm your device meets the requirements above, then check the [FAQ](/en/guide/faq).
