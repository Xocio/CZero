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
  "gc": {
    "enabled": true,
    "dirty_threshold": 200,
    "clean_threshold": 100,
    "schedule": { "every": "PT4H" },
    "script": "/data/adb/modules/CZero/list/GCclean/GCclean1",
    "wait_screen_off_timeout": 30,
    "max_runtime_sec": 600
  },
  "empty_folder": {
    "enabled": true,
    "schedule": { "every": "P1D", "at": "04:00" }
  }
}
```

Booleans use native JSON `true` / `false`.

## Field reference

### general

| Field | Type | Description |
|---|---|---|
| `auto_clean` | bool | Master switch for automatic cache cleaning |
| `log` | bool | Unified logging switch |
| `notification` | bool | Cleaning-complete notification |
| `temporal_barrier_days` | int | Temporal barrier: only clean files older than N days, `0` = disabled |

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

### gc

| Field | Type | Description |
|---|---|---|
| `enabled` | bool | GC switch |
| `dirty_threshold` | int | Dirty segments above this trigger GC (must exceed `clean_threshold`) |
| `clean_threshold` | int | Dirty segments below this count as done |
| `schedule` | schedule | GC detection frequency (minutes/hours only) |
| `script` | string | Path to the GC cleaning script |
| `wait_screen_off_timeout` | int | Screen-off wait timeout (seconds) |
| `max_runtime_sec` | int | Max GC runtime (seconds) |

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
