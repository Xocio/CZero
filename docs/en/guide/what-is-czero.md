# What is CZero

CZero is an Android root module that cleans the cache of frequently used apps, and adds background suppression, empty-folder cleanup, F2FS garbage collection and fstrim.

The module has no resident service — an extremely lightweight scheduling process triggers every task according to `config.json`, and configuration changes take effect immediately. Day-to-day operation goes through the native companion app **CZeroX**.

## The problem it solves

Over time, app caches and lingering background processes eat up storage and memory, and the device gradually slows down with less free space. CZero handles cleaning and suppression automatically in the background on a schedule, keeping the device clean and responsive without manual effort.

## Design principles

- **No resident service** — no Java process stays running; scheduling is handled by a tiny native process at near-zero cost, and cleaners exit as soon as they finish.
- **Single source of truth** — all behavior lives in one `config.json`, with no intermediate derived files.
- **Instant effect** — config changes hot-reload automatically; a broken config keeps the last-good job set and never interrupts service.
- **Native frontend** — day-to-day operation goes through the native CZeroX app, no WebUI.

## Where things live

The module and its configuration live under `/data/adb/modules/CZero/`, where `config.json` is the only file you need to care about — edit it through CZeroX day to day. Stats and logs sit in the same directory and are removed along with the module when you uninstall.

The **recycle bin** produced by cleaning is not inside the module directory; it lives in the `Recycle` folder on internal storage, where cleaned files stay recoverable for 7 days. See [Features · Recycle bin & restore](/en/guide/features#recycle-bin-restore).

## Requirements

- Android 9+ (API 28), `arm64-v8a`
- Root via Magisk, KernelSU, or APatch
- F2FS `/data` partition (only for the GC feature; everything else works regardless)

Ready? Head to [Install & Setup](/en/guide/getting-started).
