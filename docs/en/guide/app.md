# CZeroX App

**CZeroX** is CZero's native companion app, built with Jetpack Compose (Miuix style), that reads and writes the module's `config.json` and its rule JSON files directly. It is the module's only graphical configuration frontend — no WebUI.

## Download

Grab the latest APK from the [download page](/en/download) or [Releases](https://github.com/Xocio/CZero/releases). Keeping the module and the app on the same version is recommended.

## Three tabs

The app is split into **Home / Storage / Settings**.

### Home

- **Runtime status** — whether the scheduler is running and how much memory it uses; restart it in one tap if it isn't.
- **Cleaning stats** — space freed today and in total, plus execution counts for per-app cleaning, empty folders, suppression and garbage collection. Tapping through opens **Cleaning trends**, charting daily freed space over 7 / 30 days.
- **F2FS status** — live segment usage.
- **One-tap actions** — clean now, clean a single app, background suppression, F2FS garbage collection, empty-folder cleanup. Progress and results are shown via notifications, including the Dynamic Island.

### Storage

A panel unrelated to cleaning, but concerned with the same storage device:

- **Device lifespan** — shows the UFS / eMMC life level and reserved-block warning, translated into a remaining-life range.
- **Live read/write throughput** — current throughput curve and cumulative bytes.
- **Benchmark** — sequential reads/writes, random reads/writes and database transactions, combined into one score. If results vary too much across rounds, it suggests re-running.
- **Device info** — type, model, vendor, firmware and capacity.

::: tip Lifespan data is read-only
Lifespan and device info are read-only and write nothing to the storage device. The benchmark does perform real reads and writes on a test file, which is cleaned up afterwards.
:::

### Settings

Covers every field in `config.json`, grouped. Saving takes effect immediately — no reboot.

| Group | Contains |
|---|---|
| **Cleaning** | Basic cleaning, GC, background suppression (and its target app list), empty-folder cleanup (and its whitelist), enhanced cleaning, recycle bin |
| **Rules** | Custom rules (built-in rules + rule sources), cleaning whitelist, per-app cache cleaning |
| **Fine-grained suppression** | Suppression switches and capability probing, list of apps to suppress |
| **Other** | Logging, app settings (language / theme), backup & restore |

## Main pages

### Rule editor and rule sources

Manages custom cleaning paths, the cleaning whitelist and the empty-folder sweep scope. Rules are organized into groups that can be toggled as a whole and reordered.

Beyond the built-in rules, you can subscribe to third-party **rule sources**, which update once a day and appear as their own card. To publish one yourself, see [Authoring Rule Sources](/en/guide/rule-source).

### Per-app cache cleaning

Browse each path's real disk usage per app and add or remove entries; pull community-verified rules from the **shared cloud library**, or submit your own for review. Approved rules enter the public rule snapshot available to every user.

### Recycle bin

Browse cleaning records (file count, size, expiry) and restore mistakenly removed files to their original location in one tap. Cleaned files are kept 7 days by default, adjustable via `general.recycle_keep_days`. See [Features · Recycle bin & restore](/en/guide/features#recycle-bin-restore).

### Fine-grained suppression

Suppresses background apps without terminating them — state is kept, and they resume instantly. The page first probes what this device supports (system version, kernel support, the current state of the switches, and how many processes are actually suppressed), mapping to the `freezer` section of `config.json`.

A separate **apps to suppress** page lists, for each app, how many processes it has, how many are suppressed, and how much memory they hold, so you can pick per app.

::: tip How it differs from "background suppression" on Home
The one-tap background suppression on Home terminates processes; this one suspends without killing. See [Features · Fine-grained suppression](/en/guide/features#fine-grained-suppression).
:::

### Backup & restore

Exports the configuration and all rules into a single file, and imports it back in one tap after a device change or a reflash.

## Language

CZeroX supports **Simplified Chinese** and **English**, following the system setting or switchable in-app.

## Relationship to the module

CZeroX performs no cleaning itself — it only **reads and writes the configuration and displays status**; the actual scheduling and cleaning is done by the module. So the module works fine on its defaults even without CZeroX installed; the app's value is making configuration and inspection intuitive.
