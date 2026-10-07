# CZeroX App

**CZeroX** is CZero's native companion app, built with Jetpack Compose (Miuix style), that reads and writes the module's `config.json` and its rule JSON files directly. It is the module's only graphical configuration frontend — no WebUI.

## Download

Grab the latest APK from the [download page](/en/download) or [Releases](https://github.com/Xocio/CZero/releases). Keeping the module and the app on the same version is recommended.

::: info Shizuku edition
This page describes the Root edition. The Shizuku edition has no Storage tab and no fine-grained suppression, and adds a status page for the Shizuku service, authorization and the cleaning daemon; see [Choosing an Edition](/en/guide/editions).
:::

## Four tabs

The app is split into **Home / Storage / Zero / Settings**. The order and visibility of the bottom-bar items can be changed under "Settings · App settings · Theme settings · Bottom bar": long-press to drag, and switch an item off to remove it from the bar (Home and Settings are always shown).

### Home

Home is made of cards. Long-press to edit: drag to reorder, or untick a card to hide it.

- **Runtime status** — whether the scheduler is running and how much memory it uses; restart it in one tap if it isn't.
- **Cleaning stats** — space freed today and in total, plus execution counts for per-app cleaning, empty folders, suppression and garbage collection. Tapping through opens **Cleaning trends** (supporter-unlocked), charting daily freed space over 7 / 30 days.
- **Junk scan** — measures the reclaimable usage under the current cleaning rules, with per-path details.
- **File sorting** — an overview of files waiting in the source folders, with shortcuts to sort now, undo, and sorting settings.
- **F2FS status** — live segment usage.
- **Cleaning actions** — clean now, clean a single app, background suppression, F2FS garbage collection, empty-folder cleanup. Progress and results are shown via notifications, including the Dynamic Island.

### Storage

A panel unrelated to cleaning, but concerned with the same storage device:

- **Device lifespan** — shows the UFS / eMMC life level and reserved-block warning, translated into a remaining-life range.
- **Live read/write throughput** — current throughput curve and cumulative bytes.
- **Benchmark** — sequential reads/writes, random reads/writes and database transactions, combined into one score. If results vary too much across rounds, it suggests re-running.
- **Device info** — type, model, vendor, firmware and capacity.

::: tip Lifespan data is read-only
Lifespan and device info are read-only and write nothing to the storage device. The benchmark does perform real reads and writes on a test file, which is cleaned up afterwards.
:::

### Zero

The conversational assistant; see [Features · Zero assistant](/en/guide/features#zero-assistant). Configure a model service in its settings first. It is a supporter-unlocked feature and is currently in BETA.

### Settings

Covers every field in `config.json`, grouped. Changes take effect immediately — no reboot — and switch-style settings have no Save button.

| Group | Contains |
|---|---|
| **Cleaning** | Basic cleaning (including the clean interval), GC, background suppression (with its target app list and fine-grained suppression), empty-folder cleanup (and its whitelist), file sorting, recycle bin |
| **Rules** | Custom rules (built-in rules + rule sources), cleaning whitelist, per-app cache cleaning |
| **Other** | Logging, app settings (language, theme, bottom bar and more), backup & restore, uninstall |

## Main pages

### Rule editor and rule sources

Manages custom cleaning paths, the cleaning whitelist and the empty-folder sweep scope. Rules are organized into groups that can be toggled as a whole and reordered; the toolbar at the bottom of the editor supports long-press drag selection.

Beyond the built-in rules, you can subscribe to third-party **rule sources**, which update once a day and appear as their own card. To publish one yourself, see [Authoring Rule Sources](/en/guide/rule-source).

### Per-app cache cleaning

Browse each path's real disk usage per app and add or remove entries; pull community-verified rules from the **shared cloud library**, or submit your own for review. Approved rules enter the public rule snapshot available to every user.

### Recycle bin

Browse cleaning records (file count, size, expiry) and restore mistakenly removed files to their original location in one tap. Cleaned files are kept 7 days by default, adjustable via `general.recycle_keep_days`, and the bin can be switched off entirely. See [Features · Recycle bin & restore](/en/guide/features#recycle-bin-restore).

### File sorting

Manages the source folders, category folders, scan depth, sort delay and other options, and provides the sorting history: browse each sort by date and restore a whole day or only selected files. See [Features · File sorting](/en/guide/features#file-sorting).

### Fine-grained suppression

Suppresses background apps without terminating them — state is kept, and they resume instantly. The page first probes what this device supports (system version, kernel support, the current state of the switches, and how many processes are actually suppressed), and includes the **Effect check** and "Request adaptation for my device", mapping to the `freezer` section of `config.json`. It requires LSPosed with the System Framework scope ticked, and is currently in BETA.

A separate **apps to suppress** page lists, for each app, how many processes it has, how many are suppressed, and how much memory they hold, so you can pick per app.

::: tip How it differs from "background suppression" on Home
The one-tap background suppression on Home terminates processes; this one suspends without killing. See [Features · Fine-grained suppression](/en/guide/features#fine-grained-suppression).
:::

### Backup & restore

Exports the configuration and all rules into a single file, and imports it back in one tap after a device change or a reflash.

### Uninstall

Lists everything the module and the app leave on the device (configuration and stats, the recycle bin, suppression-list markers, exported files, the sorting folder and so on) and clears the ticked items in one go. Background suppression is released and the scheduler stopped first; files in the sorting folder cannot be recovered once deleted, so check before confirming.

## Appearance

Light and dark themes, a choice of Miuix or Material 3 interface style, and an optional advanced-material effect. The "Hide Settings entry" switch controls whether the app appears on the system Settings home page.

## Language

CZeroX supports **Simplified Chinese**, **English** and **Russian**, following the system setting or switchable in-app.

## Supporting the project

Some features (per-app cache cleaning, fine-grained suppression, cleaning trends, shared cloud rules, update checks and Zero) are supporter-unlocked; everything else is free. See the [FAQ](/en/guide/faq#which-features-require-a-donation).

## Relationship to the module

CZeroX performs no cleaning itself — it only **reads and writes the configuration and displays status**; the actual scheduling and cleaning is done by the module. So the module works fine on its defaults even without CZeroX installed; the app's value is making configuration and inspection intuitive.
