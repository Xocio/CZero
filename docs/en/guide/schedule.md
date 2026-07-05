# Scheduling

Every schedule in `config.json` is expressed with the same object:

```json
"schedule": { "every": "<ISO8601 duration>", "at": "<HH:MM>" }
```

- **`every`** — an ISO8601 duration meaning "how often".
- **`at`** — optional, `HH:MM` (24-hour, zero-padded), the time of day to run. **Only meaningful when the period is ≥ 1 day**.

The `timer_daemon` reads `config.json` directly and converts it into an in-memory schedule table — no intermediate derived files, and you never touch cron expressions.

## Rules for `every`

| Meaning | `every` form | `at` required |
|---|---|---|
| Every N minutes | `PT2M` (every 2 min) | No |
| Every N hours | `PT4H` (every 4 hours) | No |
| Every N days | `P1D` / `P3D` | **Yes** (specify the time) |
| Every N weeks | `P1W` / `P2W` | **Yes** |

Fixed format:

- `PT<positive int>M` / `PT<positive int>H` — time component (minutes / hours).
- `P<positive int>D` / `P<positive int>W` — date component (days / weeks).

The number must be a **positive integer ≥ 1**.

## Two kinds of schedules

| Kind | Fields | Trait |
|---|---|---|
| **Frequency** | `app_clean.detect_schedule`, `suppress.detect_schedule`, `gc.schedule` | Minute/hour intervals only, **no `at`** |
| **Timed** | `app_clean.other.schedule`, `empty_folder.schedule` | An interval (every N days/weeks) plus a time `at` |

## Examples

| `every` / `at` | Meaning |
|---|---|
| `PT2M` | Every 2 minutes |
| `PT4H` | Every 4 hours |
| `P1D` + `03:00` | Daily at 03:00 |
| `P3D` + `03:00` | Every 3 days at 03:00 |
| `P1W` + `04:00` | Once a week at 04:00 |

## Key constraints

::: warning Must follow
1. **Don't set `at` on frequency schedules** (it's ignored); **timed schedules (days/weeks) must set `at`**, otherwise they default to `00:00`.
2. **`every` must be a positive integer ≥ 1**. On a parse failure the job is skipped — that feature won't run.
3. **Changes take effect without a reboot** — use an atomic write (write a temp file, then `mv` over it).
:::

## How multi-day periods work

Cron expressions can't precisely express "every N days", so for periods ≥ 1 day the module uses **timestamp gating**: the daemon fires at the `at` time each day but, before running, checks the last-run timestamp in `cron/state` and skips if less than N days have passed. `cron/state` is maintained automatically — you only write the `schedule` object in `config.json`.
