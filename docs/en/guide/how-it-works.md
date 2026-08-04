# How It Works

This page describes how CZero operates day to day, so you have something concrete to go on when judging whether a task ran as expected.

## Overview

CZero has no resident Java service. Responsibilities are split into three stages: **install time** prepares the configuration, **boot time** starts one extremely lightweight scheduling process, and from then on that process triggers each task in the background according to the config.

```
Install  ──►  Boot  ──►  Runtime
language / inherit   start scheduler   trigger cleaning per config.json
```

The scheduler is the only resident process, and it sleeps almost all of the time, waking only to dispatch a task. The cleaners themselves exit when done and consume no background resources.

## At install time

Flashing asks you two things via the volume keys:

1. **Choose a language** — Volume Up = English, Volume Down = 中文; defaults to Chinese after roughly 60 seconds without input.
2. **Inherit the previous config** — when you inherit, your rules, whitelists and app lists carry over as-is, and `config.json` is **merged** with the new defaults: values you changed are kept and newly added options are filled in. Only a corrupt config that cannot be merged falls back to the defaults.

Once finished, the installer opens the [download page](https://czeropage.top/download-home) so the companion app is one tap away.

## After boot

Once the system is up, the module fixes permissions, initializes its state, and starts the scheduler. You can check its status on the CZeroX home page and restart it in one tap if it didn't come up.

If the [fine-grained suppression](/en/guide/features#fine-grained-suppression) switches are enabled, boot also configures the system's suppression capability once; otherwise that step is skipped entirely.

## How tasks are triggered

- **Local-time scheduling** — every task fires against the device's current time zone; you never touch cron expressions.
- **No overlap** — a round is skipped automatically if the previous one hasn't finished.
- **Missed ticks are caught up once** — tasks missed during deep sleep run once on wake; the same task is never caught up more than once, so nothing piles up into consecutive rounds.
- **Changes apply immediately** — the scheduler notices `config.json` changing and reloads on its own, no reboot needed.
- **A bad config doesn't take it down** — if the whole config fails to parse, the last valid job table is retained; if only one schedule is malformed, only that job is skipped and the rest continue.

## What one cleaning run looks like

Using WeChat cache cleaning as an example:

1. The task fires and the app's current state is checked;
2. If it is in the foreground, or the foreground is a game, the round is skipped;
3. Otherwise cleaning runs, constrained by enhanced mode and the temporal barrier;
4. Stats are updated, a log line is written if enabled, and progress and results are reported to CZeroX.

## Why this design

- **No resident service** — woken only when needed; otherwise no memory or CPU footprint.
- **Single source of truth + hot-reload** — config maps one-to-one to behavior, applies instantly, and a bad config never takes the service down.
- **Decoupled native frontend** — CZeroX only reads and writes the config files, so the module runs standalone even without the app.
