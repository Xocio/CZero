# What is CZero

CZero is an Android root module that cleans the cache of frequently used apps, and adds background suppression, empty-folder cleanup, and F2FS garbage collection.

The module has no resident service — a lightweight C++ daemon schedules every task according to `config.json`, and configuration changes take effect immediately. Day-to-day operation goes through the native companion app **CZeroX**.

## The problem it solves

Over time, app caches and lingering background processes eat up storage and memory, and the device gradually slows down with less free space. CZero handles cleaning and suppression automatically in the background on a schedule, keeping the device clean and responsive without manual effort.

## Design principles

- **No resident service** — no Java process stays running; all scheduling is handled by a tiny C++ daemon built on `timerfd` + `epoll`, at near-zero cost.
- **Single source of truth** — all behavior lives in one `config.json`, read directly by the daemon and every cleaner, with no intermediate derived files.
- **Instant effect** — the daemon watches `config.json` for changes and hot-reloads; a broken config keeps the last-good job set and never interrupts service.
- **Native frontend** — day-to-day operation goes through the native CZeroX app, no WebUI.

## Three-layer architecture

CZero is made of three cooperating layers:

| Layer | Component | Role |
|---|---|---|
| Install | `customize.sh` | Volume-key language selection, config-inheritance prompts |
| Boot | `service.sh` → `service` (C++) | Marks the module active, prepares the stats file, fixes permissions, starts the daemon |
| Runtime | `cron/timer_daemon` (C++) | Parses `config.json`, schedules and spawns all cleaning jobs, hot-reloads on change |

## Runtime files

All runtime files live under `/data/adb/modules/CZero/`:

```text
config.json            # single source of truth (read directly by daemon + cleaners)
basis/basis.prop       # cleaning statistics
cron/state             # last-run timestamps for jobs with a period >= 1 day
list/Tencent/…         # detection & cleaning scripts for frequently used apps
list/suppress/…        # background suppressor
list/GCclean/…         # F2FS GC script
list/Emptyfolder/…     # empty-folder sweeper
log/<YYYY-MM-DD>.log   # single shared daily log (only today's kept)
```

## Requirements

- Android 9+ (API 28), `arm64-v8a`
- Root via Magisk, KernelSU, or APatch
- F2FS `/data` partition (only for the GC feature; everything else works regardless)

Ready? Head to [Install & Setup](/en/guide/getting-started).
