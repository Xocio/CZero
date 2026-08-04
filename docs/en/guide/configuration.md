# Configuration

All of CZero's behavior lives in a single config file:

```
/data/adb/modules/CZero/config.json
```

It is the single source of truth, read directly by the daemon and every cleaner. Editing via [CZeroX](/en/guide/app) is recommended; when editing by hand, use an **atomic write** (write to a temp file, then replace) so the daemon never reads a half-written file.

::: tip Instant effect
Saved changes take effect **without a reboot**. The daemon watches the file and hot-reloads scheduling within a minute; non-schedule fields like enable flags and thresholds are read at runtime by each cleaner and apply immediately.
:::

## Full example

This is the default `config.json` shipped with the module:

```json
{
  "general": {
    "auto_clean": true,
    "log": false,
    "notification": false,
    "temporal_barrier_days": 3
  },
  "app_clean": {
    "detect_schedule": { "every": "PT5M" },
    "wechat": { "enabled": true, "enhanced": false },
    "qq":     { "enabled": true, "enhanced": false },
    "douyin": { "enabled": true, "enhanced": false },
    "other":  { "enabled": true, "schedule": { "every": "P1D", "at": "03:00" } }
  },
  "suppress": {
    "enabled": true,
    "detect_schedule": { "every": "PT1M" }
  },
  "freezer": {
    "auto_enable": false,
    "disable_vendor_guard": false
  },
  "gc": {
    "enabled": true,
    "free_percent": 95,
    "target_percent": 98,
    "min_interval_hours": 6,
    "require_charging": false,
    "min_battery": 40,
    "max_battery_temp": 400,
    "schedule": { "every": "PT12H" },
    "script": "/data/adb/modules/CZero/list/GCclean/GCclean1",
    "wait_screen_off_timeout": 30,
    "max_runtime_sec": 420
  },
  "trim": {
    "enabled": true,
    "schedule": { "every": "P7D", "at": "04:00" },
    "script": "/data/adb/modules/CZero/list/GCclean/GCclean1 trim"
  },
  "empty_folder": {
    "enabled": true,
    "schedule": { "every": "P1D", "at": "04:00" }
  }
}
```

Booleans use native JSON `true` / `false`.

::: tip Missing fields are harmless
Every component carries its own defaults. A single field that is absent or fails to parse **falls back to its default** without affecting the others, so you do not have to spell out every key above.
:::

## Field reference

### general

| Field | Type | Default | Description |
|---|---|---|---|
| `auto_clean` | bool | `true` | Master switch for automatic cache cleaning |
| `log` | bool | `false` | Unified logging switch |
| `notification` | bool | `false` | Cleaning-complete notification (incl. Dynamic Island) |
| `temporal_barrier_days` | int | `3` | Temporal barrier: only clean files older than N days, `0` = disabled |
| `recycle_keep_days` | int | `7` | **Optional.** Recycle-bin retention in days; `0` or absent means 7 |

### app_clean

| Field | Type | Description |
|---|---|---|
| `detect_schedule` | schedule | Frequency of foreground detection (minutes/hours only) |
| `wechat` / `qq` / `douyin` | object | Each app's `enabled` and `enhanced` (enhanced mode) flags |
| `other.enabled` | bool | Other-apps cleaning switch |
| `other.schedule` | schedule | Schedule for other-apps cleaning (multi-day + time allowed) |

### suppress

| Field | Type | Description |
|---|---|---|
| `enabled` | bool | Background suppression switch |
| `detect_schedule` | schedule | Suppression detection frequency (minutes/hours only) |

The target apps and process names are not in `config.json` — CZeroX manages them separately. See [Features · Background suppression](/en/guide/features#background-suppression).

### freezer

Boot-time actions for **fine-grained suppression**. It suppresses background apps without killing them, so their state is kept and they resume instantly — see [Features · Fine-grained suppression](/en/guide/features#fine-grained-suppression).

| Field | Type | Default | Description |
|---|---|---|---|
| `auto_enable` | bool | `false` | Enable the system's fine-grained suppression capability on every boot (Android 12+) |
| `disable_vendor_guard` | bool | `false` | Turn off the vendor's own keep-alive/suppression stack so it doesn't fight fine-grained suppression |

::: warning disable_vendor_guard has a precondition
It only acts while CZeroX's fine-grained suppression engine is enabled; otherwise it is skipped and vendor settings are left untouched.
:::

The list of apps to suppress is not in `config.json` — CZeroX manages it separately.

### gc

F2FS garbage collection. The trigger is the **free ratio**, not an absolute dirty-segment count.

| Field | Type | Default | Description |
|---|---|---|---|
| `enabled` | bool | `true` | GC switch |
| `free_percent` | int | `95` | GC only triggers when the free ratio drops **below** this; capped at 99 |
| `target_percent` | int | `98` | Considered done once the free ratio returns to this; capped at 100 and must exceed `free_percent`, otherwise it is bumped up by one point automatically |
| `min_interval_hours` | int | `6` | Minimum interval between two runs (hours), minimum 1 |
| `require_charging` | bool | `false` | When `true`, only runs while charging |
| `min_battery` | int | `40` | Minimum battery level (%) when not charging; below it, the run is skipped |
| `max_battery_temp` | int | `400` | Battery temperature ceiling in **0.1°C** (`400` = 40.0°C); minimum 100 |
| `schedule` | schedule | `PT12H` | How often to check (minutes/hours only) |
| `script` | string | — | Path to the GC program; normally no reason to change it |
| `wait_screen_off_timeout` | int | `30` | Screen-off wait timeout (seconds), minimum 5; if the screen is still on, the round is abandoned |
| `max_runtime_sec` | int | `420` | Max runtime for a single GC run (seconds), minimum 60 |

::: tip Tightening `schedule` doesn't help
`schedule` only decides how often to *check*; whether GC actually runs still has to clear `min_interval_hours`.
:::

### trim

fstrim. Scheduled **independently** of GC, with no interaction between the two.

| Field | Type | Default | Description |
|---|---|---|---|
| `enabled` | bool | `true` | fstrim switch |
| `schedule` | schedule | `P7D` + `04:00` | Schedule (multi-day + time allowed) |
| `script` | string | — | Program path and arguments; normally no reason to change it |

### empty_folder

| Field | Type | Description |
|---|---|---|
| `enabled` | bool | Empty-folder cleanup switch |
| `schedule` | schedule | Schedule (multi-day + time allowed) |

## The schedule object

Every `schedule` / `detect_schedule` uses the same shape:

```json
{ "every": "<ISO8601 duration>", "at": "<HH:MM>" }
```

`every` means "how often", `at` is an optional time-of-day (only meaningful for periods of ≥ 1 day). See [Scheduling](/en/guide/schedule).

## Config that is not in config.json

A few kinds of data are bulky or differently shaped, so they are kept out of `config.json` and managed in CZeroX instead:

- Custom cleaning paths and the cleaning whitelist (including subscribed [rule sources](/en/guide/rule-source))
- Per-app cache rules
- App lists for background suppression and fine-grained suppression
- The empty-folder sweep scope and its whitelist

Choosing to inherit during a reinstall carries all of these over.
