# How It Works

This page walks through CZero's full path from boot to cleaning, to explain *why* it's designed the way it is.

## Overview

CZero has no resident Java service. Instead it splits responsibilities into three layers: **install-time** prepares files, **boot-time** starts processes, and **runtime** is a C++ daemon that schedules every task from the config.

```
Install (customize.sh)  ──►  Boot (service.sh → service)  ──►  Runtime (timer_daemon)
   language / inherit         init / fix perms / start daemon    read config.json, schedule
```

## Install time: customize.sh

Runs once during flashing and does three main things:

1. **Language selection** — reads volume keys via `getevent`: Volume Up = English, Volume Down = 中文; defaults to Chinese after ~60s of no input. All later prompts use the chosen language.
2. **Inherit old config** (optional) — if you opt in and a prior install is detected, only the following user data is migrated; everything else resets to the new default:
   - `list/Emptyfolder/directories.prop`, `emptyfolder_white.prop`
   - `list/clean_whitelist.prop`, `list/clean_paths.prop`
   - the `basis/` stats directory
   - Note: `config.json` is **not** inherited — it always uses the new default.
3. **Opens the project page** — after install, launches the browser to the GitHub repo.

## Boot time: service.sh → service

Runs on every boot:

1. `service.sh` first fixes permissions (dirs 755; `.prop`/`.json`/`.log` etc. 644), starts the `service` binary, then `pkill`s the old `timer_daemon` and relaunches it with `setsid`.
2. `service` (C++): updates `module.prop`'s `description` to show current status, creates the `basis/basis.prop` stats file if missing, and fixes file permissions. It generates **no schedule files** — scheduling is derived entirely by the daemon reading `config.json`.

## Runtime: cron/timer_daemon

This is the core of CZero — a minimal cron engine built on `timerfd` + `epoll`:

- **Reads `config.json` directly** — a built-in JSON parser plus ISO8601→cron conversion builds the job table and gate table in memory, with no intermediate files on disk.
- **Hot-reload** — watches `config.json`'s mtime and rebuilds the job set within a minute on change; if the config is empty or unparseable, it keeps the last-good job set and never interrupts service.
- **Fixed daily job** — additionally registers a daily 00:00 `list/zero` job (reset stats, delete non-today logs) that isn't in `config.json`.
- **Minute-by-minute** — checks every minute against the phone's local time (`localtime_r`); runs scripts via `posix_spawn`; skips a round if the previous run is still in progress (overlap protection).
- **Multi-day gating** — for jobs with a period ≥ 1 day, the last-run timestamp is persisted in `cron/state`, and the run is skipped if the period hasn't elapsed.
- **Single instance** — a `cron/timer.lock` file prevents duplicate launches.
- **Logging** — writes to the shared daily log (tag `[定时]`), also gated by `general.log`.

## How a single cleaning happens

Take WeChat cache cleaning as an example:

1. The daemon triggers `list/Tencent/check` on schedule.
2. `check` identifies the foreground app: reads `/dev/cpuset/top-app` and `dumpsys activity`'s `ResumedActivity`.
3. It evaluates safeguards: if the foreground is WeChat itself, or a game → skip this round; otherwise continue.
4. It invokes the matching cleaner (`tencentmm` etc., a Go binary), constrained by `enhanced` and the temporal barrier.
5. It updates `basis.prop` stats, writes a log if enabled, and can broadcast progress / result to CZeroX.

## Why this design

- **No resident service** — woken by the daemon only when needed; otherwise no memory or CPU footprint.
- **Single source of truth + hot-reload** — config maps one-to-one to behavior, applies instantly, and a bad config never takes the service down.
- **Decoupled native frontend** — CZeroX only reads/writes `config.json` and is fully decoupled from the module, which runs standalone even without the app.
