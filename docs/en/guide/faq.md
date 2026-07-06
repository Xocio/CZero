# FAQ

## Do I need to reboot after changing the config?

No. The daemon watches `config.json` and hot-reloads scheduling within a minute; non-schedule fields like enable flags and thresholds are read at runtime and apply immediately. When editing by hand, use an atomic write (temp file, then replace).

## My device isn't F2FS — can I still use it?

Yes. Every feature except **F2FS GC** is independent of the filesystem type. On non-F2FS devices, set `gc.enabled` to `false`; everything else works normally.

## Can I use it without CZeroX?

Yes. The module runs fine on the default `config.json`. CZeroX just makes configuring and monitoring intuitive; without it you edit `config.json` by hand.

## Which root solutions are supported?

Magisk, KernelSU, and APatch.

## Will reinstalling lose my settings?

At install you can choose to inherit the old config, which carries over the black/white lists and custom path list; but `config.json` is always reset to the new default, so re-adjust it after reinstalling if needed.

## What happens if I write a broken config?

If the whole `config.json` fails to parse, the daemon **keeps the last-good job set** and never interrupts service. If only one job's `every` is invalid, **only that job** is skipped; the rest keep running.

## Where are the logs?

With `general.log` on, logs are at `/data/adb/modules/CZero/log/<YYYY-MM-DD>.log`, keeping only today's file. Each line is tagged with its source component.

## Could cleaning delete my data?

Cleaning targets only caches and leftover directories, with two layers of protection:

- **Temporal barrier** — governed by `temporal_barrier_days`, only files older than N days are touched; for a more conservative setup, raise that value or narrow the scope.
- **Recycle bin** — other / custom-path cleaning **does not truly delete**. Files are first moved to the recycle bin and kept for **7 days** by default, fully recoverable to their original location during that window. So even an accidental removal can be undone. See [Features · Recycle bin & restore](/en/guide/features#recycle-bin-restore).

## I deleted a file by mistake — how do I restore it?

Cleaned files stay in the recycle bin for 7 days. View and restore them in CZeroX.

Restoring moves files back to their original paths as-is (preserving their original mtime). Sessions older than 7 days are purged automatically and can no longer be restored.

## Where do I report issues?

Please open one on [GitHub Issues](https://github.com/Xocio/CZero/issues).
