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
"wechat": { "enabled": true, "enhanced": false }
```

- `enabled` — whether to clean this app.
- `enhanced` — **enhanced mode**. Widens the cleaning scope; off by default for safety.

### Other-apps cleaning

`app_clean.other` handles caches beyond the big three, run on its `schedule` (default daily 03:00).

### Master switch & temporal barrier

- Governed by the master switch `general.auto_clean`.
- Protected by `general.temporal_barrier_days` (the **temporal barrier**): only files older than N days are cleaned, protecting recent data; `0` disables it.

## Background suppression

At the frequency of `suppress.detect_schedule`, terminates the target apps' background processes to free memory and reduce background drain.

- **Subprocesses only** — the main process is untouched, so notifications and push still work.
- **Never touches the foreground app** — an app currently in the foreground is never suppressed, and if the foreground app cannot be determined this round, suppression is skipped entirely.
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

`max_runtime_sec` additionally caps a single run and force-stops it when reached.

::: tip Rarely seeing it run is normal
The admission conditions are deliberately strict, and CZero yields when the system is already performing its own maintenance. If several runs in a row achieve little, the interval lengthens automatically to avoid pointless retries.
:::

> GC requires an F2FS `/data` partition. On non-F2FS devices, set `gc.enabled` to `false`; everything else works regardless.

## fstrim

`trim` is a separate job from GC (weekly at 04:00 by default). It tells the underlying storage which space has been freed, so the UFS / eMMC controller's own reclamation has something to work with — which keeps write performance healthy over time.

- Scheduled **independently** of GC and unaffected by `gc.enabled`.
- Also requires the screen to be off; the round is skipped if the screen is on.

::: warning Don't run it too often
fstrim puts real write pressure on the storage device. The weekly default is plenty; daily is not recommended.
:::

## Fine-grained suppression

[Background suppression](#background-suppression) above simply terminates background processes. **Fine-grained suppression** is the gentle counterpart — it **suppresses without killing**: a suppressed app stops burning CPU, but its in-memory state is preserved, so switching back to it resumes instantly with nothing to reload.

The capability comes from Android itself. The module does not perform it — it only brings it into a usable state at boot:

| Field | What it does |
|---|---|
| `auto_enable` | Enables the system's fine-grained suppression capability on each boot |
| `disable_vendor_guard` | Turns off the vendor's own keep-alive/suppression stack so the two strategies don't fight |

The list of apps to suppress is managed in CZeroX, not written into `config.json`.

::: tip Which one to use
They coexist. Background suppression suits subprocesses that are harmless to kill and only waste memory otherwise; fine-grained suppression suits apps whose live state you want preserved but which shouldn't be active in the background.
:::

::: warning Requires Android 12+
On older versions these two switches have no effect.
:::

## Empty-folder cleanup

On the `empty_folder.schedule` (default daily 04:00), sweeps away leftover empty directories. Both the sweep scope and its whitelist are editable in CZeroX, organized into groups that can be toggled as a whole.

## Custom-path cleaning

Beyond the built-in rules, you can declare your own cleaning paths for the daily job, and use a whitelist to protect directories that should never be touched. Both are managed in CZeroX's rule editor and are inherited across reinstalls.

Rules support wildcards and are deduplicated automatically (paths beneath a directory already on the list are not processed twice); the whitelist always takes priority.

Alongside the path-grouped rules there is a second set organized **by app**: the "App cache" page in CZeroX shows the real disk usage of each path per app, lets you add or remove entries, and can pull community-verified rules from the shared cloud library. Both sets are merged at cleaning time and behave identically.

You can also subscribe to third-party [rule sources](/en/guide/rule-source), which update once a day.

Cleaning does not delete outright — matched files are first moved to the recycle bin (see below), kept for 7 days by default, and fully recoverable during that window.

## Recycle bin & restore

To make accidental deletion recoverable, other / custom-path cleaning **does not truly delete**. Instead it **moves** matched files into a recycle bin, keeping them for a while and only purging them once they expire.

- **Move, not delete** — cleaned files retain their original directory structure and can be **restored to their exact original location**, symlinks included.
- **Archived per run** — each cleaning run forms its own record, showing which files it removed, how large they were, and when they expire.
- **Kept for 7 days by default** — adjustable via `general.recycle_keep_days` (absent or `0` means 7). Expired records are purged automatically.
- **Hidden from the gallery** — moved photos and videos don't show up in the media library.
- **Preserves mtime** — restored files aren't misjudged as "new" by the temporal barrier.
- **Self-protection** — even if a rule matches the root of internal storage, the recycle bin itself is never cleaned.

The recycle bin lives in the `Recycle` folder on internal storage; day to day, browse and restore from CZeroX's recycle-bin page.

## Stats & logging

- **Stats** — per-app cleaning counts, suppression counts and freed space are accumulated, with a 90-day daily history that powers CZeroX's "Cleaning trends" chart. Counters roll over at 00:00 daily, and a device that was off at midnight does not miss it.
- **Logging** — with `general.log` on, all components write to one shared daily log, each line tagged with its source (`[定时]` `[检测]` `[微信]` `[压制]` `[GC]` `[空文件夹]` `[自定义]` and so on) so you can tell which stage produced it. Only the current day is kept.

Both are off by default and can be toggled at any time.

## Dynamic Island notifications

Cleaning start and finish are reported to CZeroX, which displays progress and results (including the Dynamic Island). Controlled by `general.notification`.
