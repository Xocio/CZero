# Features

Every CZero feature can be toggled independently in `config.json` and is triggered by the daemon on its own schedule. This page explains **what each feature does, how it's triggered, and what safeguards apply**.

## Cache cleaning

Dedicated cleaners for frequently used apps (WeChat, QQ, Douyin), triggered at the frequency of `app_clean.detect_schedule`.

### Checks before triggering

On each trigger, `check` first inspects the app's current state to avoid pointless or harmful cleaning:

- **Only cleans apps that should be cleaned** — identifies the foreground app by reading `/dev/cpuset/top-app` and `dumpsys activity` (`ResumedActivity`), using `oom_score_adj` to pin down the foreground main process.
- **Never cleans the foreground app** — if an app is currently in use in the foreground, it's skipped this round so ongoing actions aren't disrupted.
- **Game-foreground protection** — a built-in list of common game packages; when a game is in the foreground, cleaning is skipped so gameplay isn't interrupted.

### Per-app switches

```json
"wechat": { "enabled": true, "enhanced": false }
```

- `enabled` — whether to clean this app.
- `enhanced` — **enhanced mode**. When on, cleaning is more thorough (wider scope); off by default for safety.

### Other-apps cleaning

`app_clean.other` handles caches beyond the big three, run on its `schedule` (default daily 03:00).

### Master switch & temporal barrier

- Governed by the master switch `general.auto_clean`.
- Protected by `general.temporal_barrier_days` (the **temporal barrier**): only files older than N days are cleaned, protecting recent data; `0` disables it.

## Background suppression

At the frequency of `suppress.detect_schedule`, periodically suppresses target apps running in the background.

How it works: it walks `/proc`, finds the target apps' background processes, and terminates them (`SIGKILL`), freeing memory and reducing background drain. Key safeguards:

- **Never touches the foreground app** — an app currently in the foreground is never suppressed.
- **No suppression if the foreground is unknown** — if this round can't determine the foreground app, it's skipped to avoid collateral damage.
- Suppression counts are recorded in the stats (`suppress` in `basis.prop`).

## F2FS garbage collection

Monitors the F2FS dirty-segment count and runs garbage collection (GC) when it exceeds a threshold. For safety, GC waits for the screen to turn off before running and caps its runtime.

| Field | Meaning |
|---|---|
| `dirty_threshold` | GC triggers when dirty segments exceed this (must be greater than `clean_threshold`) |
| `clean_threshold` | Considered done and stops when dirty segments drop below this |
| `wait_screen_off_timeout` | Timeout (seconds) waiting for screen-off; if still on, this round is abandoned |
| `max_runtime_sec` | Max runtime (seconds) for a single GC run; force-stopped when reached |

::: tip Why wait for screen-off
F2FS GC is disk-intensive; running it while the screen is off and the device is idle avoids affecting foreground responsiveness.
:::

> GC requires an F2FS `/data` partition. On non-F2FS devices, set `gc.enabled` to `false`; everything else works regardless.

## Empty-folder cleanup

On the `empty_folder.schedule` (default daily 04:00), sweeps away leftover empty directories. The scope is defined by `list/Emptyfolder/directories.prop` and constrained by the whitelist `emptyfolder_white.prop`.

## Custom-path cleaning

Beyond the built-in rules, you can declare your own paths, swept by the daily job:

| File | Purpose |
|---|---|
| `list/clean_paths.prop` | List of custom paths to clean |
| `list/clean_whitelist.prop` | Cleaning whitelist (paths protected from cleaning) |

These lists can be managed via the CZeroX rule editor and are inherited across reinstalls.

Cleaning does not delete outright — matched files are first moved to the recycle bin (see below), kept for 7 days by default, and fully recoverable during that window.

## Recycle bin & restore

To make accidental deletion recoverable, other / custom-path cleaning **does not truly delete**. Instead it **moves** matched files into a recycle bin, keeping them for a while and only purging them once they expire.

### How it works

- **Move, not delete** — each cleaned file is mirrored into the recycle bin under its original directory structure, so it can be **restored to its exact original location**.
- **Archived per session** — each cleaning run creates its own session directory named `<timestamp>-<PID>` (e.g. `20060102-150405-12345`), plus a `manifest.json` recording every file's **original path, size, creation time, and expiry time**.
- **Kept for 7 days** — the retention period is fixed at **7 days** (not adjustable via `config.json`). Expired sessions are purged automatically on the next cleaning run (or manually via `purge`).
- **No empty directories** — if a run moves nothing, no session directory is left on disk; after a restore completes, its session directory is removed too.
- **Hidden from the gallery** — a `.nomedia` file is written at the recycle root so moved photos / videos don't show up in the media library.
- **Preserves mtime** — both move and restore keep each file's original mtime, so restored files aren't misjudged as "new" by the temporal barrier.
- **Self-protection** — even if a cleaning rule matches `/storage/emulated/0`, the recycle bin itself is never cleaned, avoiding a recursive loop.

The recycle bin lives on internal storage:

```text
/storage/emulated/0/Recycle/<session>/files/<mirrored original path>
/storage/emulated/0/Recycle/<session>/manifest.json
```

### Management commands

The recycle bin is served by the `list/customize` binary through three subcommands, for CZeroX or manual use (all output JSON for easy parsing):

| Command | Purpose |
|---|---|
| `customize list` | List all recycle sessions with their manifests (file count, size, expiry) for display |
| `customize restore <session-id>` | Move all files in the given session back to their original locations |
| `customize restore all` | Restore every session |
| `customize purge` | Immediately clear all expired recycle sessions |

::: tip Recoverable even without a manifest
If the cleaning process was interrupted and `manifest.json` is missing, `list` / restore rebuild the manifest by scanning the session directory on the fly, so restoring still works.
:::

## Stats & logging

- **Stats** — per-app cleaning counts and suppression counts are recorded in `basis/basis.prop`; the CZeroX home page shows today's and total figures. The `zero` job resets the day's counters at 00:00.
- **Logging** — with `general.log` on, all components write to a single daily log `log/<YYYY-MM-DD>.log`, each line tagged with its source:

  | Tag | Source |
  |---|---|
  | `[定时]` | Scheduling daemon |
  | `[检测]` | Foreground detection |
  | `[微信]` `[QQ]` `[抖音]` | Per-app cleaning |
  | `[压制]` | Background suppression |
  | `[GC]` | F2FS garbage collection |
  | `[空文件夹]` | Empty-folder cleanup |
  | `[自定义]` | Custom-path cleaning |

  Only today's log is kept (the `zero` job deletes non-today log files).

Both are off by default and can be toggled at any time.

## Dynamic Island notifications

On start and finish, cleaners can broadcast events to CZeroX (`CleanEventReceiver`) to show progress / result notifications (including the Dynamic Island). This can be disabled via the environment variable `CZERO_NO_NOTIFY=1`.
