# Features

Every CZero feature can be toggled independently in `config.json` and runs on its own schedule. This page covers **what each feature does, when it runs, and what safeguards apply**. For field-by-field values, see [Configuration](/en/guide/configuration).

## Cache cleaning

Dedicated cleaners for frequently used apps (WeChat, QQ, Douyin), triggered at the frequency of `app_clean.detect_schedule`.

### Checks before triggering

Each trigger first examines the app's current state to avoid pointless or harmful cleaning:

- **Never cleans the foreground app** — if an app is currently in use in the foreground, it is skipped this round so ongoing actions aren't disrupted.
- **Game-foreground protection** — cleaning is skipped while a game is in the foreground.

### Per-app switches

```json
"wechat": { "enabled": true }
```

- `enabled` — whether to clean this app.

Cleaning for WeChat, QQ and Douyin also covers data under dual apps (app clones).

::: tip Enhanced mode has been removed
The "enhanced mode" of earlier versions is retired. A leftover `enhanced` field in an old config no longer has any effect and is removed automatically after upgrading.
:::

### Clean interval

`app_clean.min_interval_hours` (default 12 hours) sets the **minimum interval between two automatic cleans of the same app**. The detection frequency only decides how often the state is checked; a clean also has to clear this interval, so switching between apps repeatedly does not trigger repeated cleaning, which avoids needless battery drain and stutter. Changes take effect immediately.

### Other-apps cleaning

`app_clean.other` handles caches beyond the big three, run on its `schedule` (default daily 03:00).

### Master switch & temporal barrier

- Governed by the master switch `general.auto_clean`.
- Protected by `general.temporal_barrier_days` (the **temporal barrier**): only files older than N days are cleaned, protecting recent data; `0` disables it.

## Background suppression

At the frequency of `suppress.detect_schedule`, terminates the target apps' background processes to free memory and reduce background drain.

- **Subprocesses only** — the main process is untouched, so notifications and push still work.
- **Never touches the foreground app** — an app currently in the foreground is never suppressed, and if the foreground app cannot be determined this round, suppression is skipped entirely.
- **Low overhead** — detection does not relaunch a process every round and does not query window or power state frequently, so the impact on smoothness and battery life is small.
- **Editable list** — WeChat / QQ / Alipay by default; apps can be added, removed, or toggled individually in CZeroX.

## F2FS garbage collection

Runs F2FS garbage collection (GC) when free space gets tight, consolidating fragmented space.

GC is disk-intensive, so it only acts when the device is genuinely idle. Failing any of the following skips the round:

| Condition | Field | Default |
|---|---|---|
| Storage actually needs reclaiming | `free_percent` / `target_percent` | triggers below 95% free, considered done at 98% |
| Screen must be off | `wait_screen_off_timeout` | waits up to 30s, then gives up |
| Interval since the last run | `min_interval_hours` | 6 hours |
| Must be charging | `require_charging` | off (not required) |
| Minimum battery when not charging | `min_battery` | 40% |
| Battery temperature ceiling | `max_battery_temp` | 40.0°C |

`max_runtime_sec` additionally caps a single run and force-stops it when reached. If the screen turns on during a run, heavy reclamation stops immediately.

::: tip Rarely seeing it run is normal
The admission conditions are deliberately strict, and CZero yields when the system is already performing its own maintenance. If several runs in a row achieve little, the interval lengthens automatically to avoid pointless retries.
:::

> GC requires an F2FS `/data` partition. On non-F2FS devices, set `gc.enabled` to `false`; everything else works regardless.

::: tip Shizuku edition
The Shizuku edition has no module-provided reclamation method; the system's idle maintenance does the work instead, so the conditions in the table above do not apply and have no settings. See [Choosing an Edition](/en/guide/editions).
:::

## fstrim

`trim` is a separate job from GC (weekly at 04:00 by default). It tells the underlying storage which space has been freed, so the UFS / eMMC controller's own reclamation has something to work with — which keeps write performance healthy over time.

- Scheduled **independently** of GC and unaffected by `gc.enabled`.
- Also requires the screen to be off; the round is skipped if the screen is on.

::: warning Don't run it too often
fstrim puts real write pressure on the storage device. The weekly default is plenty; daily is not recommended.
:::

## Fine-grained suppression

::: info Root edition only
This feature relies on the system framework and is not available in the Shizuku edition.
:::

[Background suppression](#background-suppression) above simply terminates background processes. **Fine-grained suppression** is the gentle counterpart — it **suppresses without killing**: once in the background, an app is suspended and neither uses CPU nor runs background tasks, yet its in-memory state is preserved, so switching back resumes instantly with nothing to reload.

::: warning BETA
This feature is currently in BETA and may fail to take effect on some devices. If abnormal system behavior is detected during the previous activation, it is switched off automatically with a notice; simply turn it on again to continue.
:::

### Prerequisites

- Android 12 or later, with kernel support for app freezing. On older versions the feature has no effect.
- LSPosed installed, with the **System Framework** scope ticked for CZeroX. Without it the related switches do nothing.
- A single reboot after enabling before it takes effect.

### Suppression list

On first use a default set of apps is included (instant-messaging, music and navigation apps excluded); add or remove apps in CZeroX's suppression list as needed. A suppressed app cannot receive push notifications, so instant-messaging, music and navigation apps are not recommended. System-critical apps such as the keyboard, launcher, dialer and messaging are always excluded and cannot be added.

### Effect check

CZeroX's **Effect check** page actually freezes a test process and checks, item by item, whether the framework hook, the freezer and vendor restrictions are really working, so you can tell whether the feature is usable on your device. If the check fails, "Request adaptation for my device" packages device information to send to the developer.

### Vendor restrictions & boot actions

Vendor freezing restrictions on Android 17 and HyperOS 4 are accommodated. The `freezer` section of `config.json` controls the related boot-time actions:

| Field | What it does |
|---|---|
| `auto_enable` | Enables the system's fine-grained suppression capability on each boot |
| `disable_vendor_guard` | Turns off the vendor's own keep-alive/suppression stack so the two strategies don't fight |

The list of apps to suppress is managed in CZeroX, not written into `config.json`.

::: tip Which one to use
They coexist. Background suppression suits subprocesses that are harmless to kill and only waste memory otherwise; fine-grained suppression suits apps whose live state you want preserved but which shouldn't be active in the background.
:::

## Empty-folder cleanup

On the `empty_folder.schedule` (default daily 04:00), sweeps away leftover empty directories. Both the sweep scope and its whitelist are editable in CZeroX, organized into groups that can be toggled as a whole.

Safeguards during the sweep:

- Directories the system creates for apps are never removed, so apps do not fail to write files afterwards;
- Other mount points and symbolic links are never entered;
- WeChat and Douyin are on the protection whitelist by default.

## Custom-path cleaning

Beyond the built-in rules, you can declare your own cleaning paths for the daily job, and use a whitelist to protect directories that should never be touched. Both are managed in CZeroX's rule editor and are inherited across reinstalls.

Rules support wildcards and are deduplicated automatically (paths beneath a directory already on the list are not processed twice); the whitelist always takes priority. Whitelist entries match on whole path components, so a short, generic entry does not protect other folders with similar names; the whitelist's "Select apps" page also adds an app in one tap.

Alongside the path-grouped rules there is a second set organized **by app**: the "App cache" page in CZeroX shows the real disk usage of each path per app, lets you add or remove entries, and can pull community-verified rules from the shared cloud library. Both sets are merged at cleaning time and behave identically.

You can also subscribe to third-party [rule sources](/en/guide/rule-source), which update once a day.

Cleaning does not delete outright — matched files are first moved to the recycle bin (see below), kept for 7 days by default, and fully recoverable during that window.

## Recycle bin & restore

To make accidental deletion recoverable, other / custom-path cleaning **does not truly delete**. Instead it **moves** matched files into a recycle bin, keeping them for a while and only purging them once they expire.

- **Move, not delete** — cleaned files retain their original directory structure and can be **restored to their exact original location**, symlinks included.
- **Archived per run** — each cleaning run forms its own record, showing which files it removed, how large they were, and when they expire.
- **Kept for 7 days by default** — adjustable via `general.recycle_keep_days` (absent or `0` means 7), affecting only records created afterwards. Expired records are purged automatically.
- **Can be turned off** — with `general.recycle_enabled` or CZeroX's recycle-bin settings switched off, cleaning deletes permanently and cannot be undone; use with care.
- **Hidden from the gallery** — moved photos and videos don't show up in the media library.
- **Preserves mtime** — restored files aren't misjudged as "new" by the temporal barrier.
- **Self-protection** — even if a rule matches the root of internal storage, the recycle bin itself is never cleaned.

The recycle bin lives in the `Recycle` folder on internal storage; day to day, browse and restore from CZeroX's recycle-bin page.

## File sorting

Automatically tidies files scattered across download folders and similar locations, moving them into category folders by type. Off by default; turn it on from CZeroX's file-sorting page, or run "Sort now" manually at any time.

- **Source folders** — the folders to tidy. Common locations (downloads, browsers and download managers, chat apps, cloud drives, photos and media) can be added in one tap; higher-risk locations are marked "Caution".
- **Destination & categories** — files go to the `CZero` folder on internal storage by default, split into packages, archives, documents, images, videos, audio, scripts, fonts and others. Each category's folder name is customizable, a category can be set to "Don't sort" to leave its files in place, and the language of all folder names can be switched in one tap.
- **Scan depth** — handle only the source folder itself, a number of levels below it, or unlimited; the folders themselves are never moved.
- **Sort delay** — a file must sit untouched for a while after it is written (60 seconds by default) before it is moved, so files that are still settling are left alone; files arriving in succession are handled together.
- **Duplicates** — same-name, same-size files can go to a separate duplicates folder instead of getting a numeric suffix.
- **Package renaming** — installers can be renamed after the app they contain once sorted.
- **Automatic skips** — unfinished downloads (`.crdownload`, `.part`, etc.) and freshly written files are not touched; unfinished downloads are listed separately in the overview and can be safely deleted.
- **Fully reversible** — the sorting history records every move by date; restore a whole day or only selected files. File lists support thumbnail previews, direct sharing and second-level timestamps.

Corresponds to the `file_sort` section of `config.json`; see [Configuration](/en/guide/configuration#file-sort) for the fields.

::: tip Edition differences
File sorting is generally available and free in the Root edition; in the Shizuku edition it is still in BETA and supporter-unlocked.
:::

For a first run, start with the download folder and review the files listed in the "Confirm sorting" preview.

## Junk scan

The junk-scan card on CZeroX's home page measures the real disk usage of every path under the **currently active cleaning rules** and shows the total reclaimable size with a per-item breakdown. Scanning is read-only and deletes nothing.

## Zero assistant

Zero is a conversational assistant inside CZeroX. You provide your own model-service API key (OpenAI-compatible and Anthropic-compatible endpoints are supported). It can scan your storage, preview how much a cleanup would free, browse folders and turn them into custom rules, manage app suppression, check module health and read your cleaning history.

- **Confirmation first** — anything that changes the device is presented as a confirmation card and only runs once you confirm.
- **Keeps running in the background** — a conversation continues after you leave the app; progress shows in a notification (and Dynamic Island), where you can confirm or cancel.
- **Sessions** — conversations are saved as separate, auto-titled sessions that you can revisit or delete.

This is a supporter-unlocked feature and is currently in BETA.

## Stats & logging

- **Stats** — per-app cleaning counts, suppression counts and freed space are accumulated, with a 90-day daily history that powers CZeroX's "Cleaning trends" chart (supporter-unlocked). Counters roll over at 00:00 daily, and a device that was off at midnight does not miss it.
- **Logging** — with `general.log` on, all components write to one shared daily log, each line tagged with its source (`[定时]` `[检测]` `[微信]` `[压制]` `[GC]` `[空文件夹]` `[自定义]` `[归类]` and so on) so you can tell which stage produced it. Only the current day is kept.

Both are off by default and can be toggled at any time.

## Dynamic Island notifications

Cleaning start and finish are reported to CZeroX, which displays progress and results (including the Dynamic Island). Controlled by `general.notification`; notifications from empty-folder cleaning and checks honor the same switch.
