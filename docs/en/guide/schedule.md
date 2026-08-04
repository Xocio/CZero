# Scheduling

Every schedule in `config.json` is expressed with the same object:

```json
"schedule": { "every": "<ISO8601 duration>", "at": "<HH:MM>" }
```

- **`every`** — an ISO8601 duration meaning "how often".
- **`at`** — optional, `HH:MM` (24-hour, zero-padded), the time of day to run. **Only meaningful when the period is ≥ 1 day**.

The module reads `config.json` directly and derives the schedule from it — no intermediate derived files, and you never touch cron expressions.

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

The number must be a **positive integer ≥ 1**, with minutes no greater than **59** and hours no greater than **23** — they end up in a per-minute / per-hour bitmap, so a form like `PT90M` cannot be expressed and is rejected as an invalid schedule. For "every 90 minutes", use `PT1H` or `PT2H` instead.

## Two kinds of schedules

| Kind | Fields | Trait |
|---|---|---|
| **Frequency** | `app_clean.detect_schedule`, `suppress.detect_schedule`, `gc.schedule` | Minute/hour intervals only, **no `at`** |
| **Timed** | `app_clean.other.schedule`, `empty_folder.schedule`, `trim.schedule` | An interval (every N days/weeks) plus a time `at` |

::: tip Intervals are wall-clock aligned, not "since last run"
Minute / hour intervals snap to the clock: `PT4H` fires at 00, 04, 08, 12, 16 and 20 — not "4 hours after the last run". Minutes work the same way, so `PT5M` lands on :00, :05, :10 and so on.
:::

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
2. **`every` must be a positive integer ≥ 1**, with minutes ≤ 59 and hours ≤ 23. On a parse failure the job is skipped — that feature won't run, and the log records an "invalid schedule" line.
3. **Changes take effect without a reboot** — use an atomic write (write a temp file, then `mv` over it).
:::

## A few notes

- **Multi-day periods are measured by actual elapsed time**, not "fire whenever the clock hits the time". A shutdown or a missed time-of-day does not drop the task; it runs at the next occurrence.
- **Tasks missed during sleep are caught up once** — never repeated for every hour that was missed.
- **`gc.schedule` is not GC's run frequency.** It only decides how often to check; actually running still requires `gc.min_interval_hours` and the other admission conditions, so tightening it does not make GC run more often.
- **Stats rollover and log cleanup** happen at 00:00 daily and are not controlled by `config.json`.
