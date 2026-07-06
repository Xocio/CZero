# CZeroX App

**CZeroX** is CZero's native companion app, built with Jetpack Compose, that reads and writes the module's `config.json` directly. It is the module's only graphical configuration frontend — no WebUI.

## Download

Download the latest CZeroX APK from the [Releases](https://github.com/Xocio/CZero/releases) page and install it.

## Key capabilities

### Home dashboard

- **Daemon status** — PID and memory usage; restart it in one tap if it isn't running.
- **Cleaning stats** — today's and total figures (from `basis.prop`).
- **F2FS status card** — live dirty / free / used / total segment counts.

### One-tap actions

- Clean now (run all cleaning jobs)
- Per-app cleaning
- Background suppression
- F2FS GC (primary / backup plan)
- Empty-folder cleanup

Progress and results are surfaced via notifications.

### Settings editor

Covers every field in `config.json`: feature switches, schedule intervals and times, GC thresholds, the temporal barrier, and more. Saved changes are hot-reloaded by the daemon — no reboot.

### Rule editor

Manage the black/white lists and the custom path list (`clean_paths.prop`) to widen or narrow the cleaning scope.

### Recycle bin

View the recycle sessions produced by cleaning (file count, size, expiry) and restore accidentally deleted files back to their original locations with one tap. The data comes from the `list` / `restore` subcommands of `list/customize`; cleaned files are kept for 7 days by default. See [Features · Recycle bin & restore](/en/guide/features#recycle-bin-restore).

## Language

CZeroX supports **English** and **简体中文**, following the system or switchable in-app.

## Relationship to the module

CZeroX does no cleaning itself — it only **reads/writes `config.json` and displays status**; the actual scheduling and cleaning are done by the C++ daemon inside the module. So even without CZeroX the module runs fine on its default config — the app simply makes configuring and monitoring intuitive.
