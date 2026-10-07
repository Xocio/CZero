# FAQ

## Is any setup required after installing?

No. The default configuration already enables cache cleaning, background suppression, garbage collection and empty-folder cleanup; after a reboot it runs in the background on schedule. CZeroX is for viewing status and adjusting settings — it is not required.

## Do I need to reboot after changing a setting?

No. Saving is enough.

## Which root solutions are supported?

Magisk, KernelSU and APatch. The device must run Android 9 or newer on `arm64`.

## Can I use it without CZeroX?

Yes, the module runs independently. However, there is then no graphical interface and settings must be edited by hand in the config file, which is not recommended.

## Could cleaning remove data I care about?

Cleaning targets only caches and leftover directories, with two layers of protection:

- **Temporal barrier** — by default only files older than 3 days are processed, leaving recent data untouched. Raise the value for a more conservative setup.
- **Recycle bin** — cleaning does not delete outright; files are first moved to the recycle bin and kept for 7 days, fully recoverable during that window.

## How do I restore a file removed by mistake?

Find the record in CZeroX's **Recycle bin** and restore it; the file is moved back to its original location as it was. Records past the retention period are purged automatically and can no longer be restored.

The retention period can be raised in the recycle-bin settings, at the cost of more storage used.

## My device isn't F2FS — can I still use it?

Yes. Everything except **F2FS garbage collection** is independent of the filesystem type. On a non-F2FS device, simply turn that feature off.

## Garbage collection never seems to run?

This is expected. Garbage collection does not run merely because it is scheduled; all of the following must hold:

- storage genuinely needs reclaiming (nothing to do when free space is ample);
- enough time has passed since the last run;
- the screen is off;
- the battery has sufficient charge and a normal temperature.

Consequently, never observing it run usually indicates the device is in good shape. To confirm, enable logging — the log records which condition caused the round to be skipped.

## How does fstrim differ from garbage collection?

**Garbage collection** consolidates fragmentation within the device's filesystem; **fstrim** informs the storage device which space has been freed, so its own reclamation has something to work with. The two are complementary, both enabled by default, and require no further configuration.

## How does fine-grained suppression differ from background suppression?

- **Background suppression** — terminates background processes outright, freeing memory immediately at the cost of a full reload next launch.
- **Fine-grained suppression** — suspends the app without terminating it, stopping CPU usage while preserving its full runtime state for instant resumption.

Use the former to reclaim resources thoroughly, the latter to preserve the app's live state. Both can be enabled at once. Fine-grained suppression requires Android 12 or newer, plus LSPosed with the System Framework scope ticked for CZeroX.

## Fine-grained suppression isn't taking effect. What should I do?

First confirm in CZeroX's fine-grained suppression page that it is switched on and that you have rebooted, then open **Effect check**, which tests the framework hook, the freezer and vendor restrictions item by item and shows which one fails. The usual causes are the System Framework scope not being ticked for CZeroX in LSPosed, or a kernel without app-freezing support. If it still doesn't work, use "Request adaptation for my device" to package the information and send it to the developer.

## What is the difference between the Root and Shizuku editions?

Cleaning rules and most features are the same; the difference is how privileges are obtained and the feature set that follows. The Shizuku edition needs no root or module, but has no fine-grained suppression or Storage tab, and F2FS reclamation is done by the system's idle maintenance. See [Choosing an Edition](/en/guide/editions).

## The Shizuku edition stopped cleaning after a reboot

Shizuku started via ADB stops on every reboot and has to be started once more before scheduled cleaning resumes; Shizuku started via Root is available automatically after boot. If Shizuku is running but tasks still do not run, tap "Redeploy" on the app's status page.

## What does file sorting move?

Only files inside the source folders you added, down to the configured scan depth, are moved into the sorting folder by type. Unfinished downloads and freshly written files are skipped, and the folders themselves are never moved. Every move is recorded in the sorting history, where you can restore a whole day or only selected files. The feature is off by default. It is free in the Root edition; in the Shizuku edition it is still BETA and supporter-unlocked.

## Which features require a donation?

Per-app cache cleaning, fine-grained suppression, cleaning trends and daily history, shared cloud rules, app and module update checks, and the Zero assistant are supporter-unlocked. WeChat / QQ / Douyin cleaning, custom rules and whitelist, F2FS reclamation, empty-folder cleanup, basic background suppression and the recycle bin are free; file sorting is free in the Root edition and supporter-unlocked in the Shizuku edition. For an activated order, the bound devices can be viewed in Device management; unbinding an old device after changing phones frees its slot.

## Will reinstalling or updating lose my settings?

No. Choose **inherit** when flashing: existing rules, whitelists and adjusted options are all retained, and options added by the new version are filled in automatically.

## Does uninstalling leave files behind?

No. Remove the module in your root manager and reboot; the module directory is cleared along with its statistics and logs. The recycle bin resides in the `Recycle` folder on internal storage and can be deleted manually for a complete cleanup.

## Where can I view the logs?

Enable logging in CZeroX and logs are stored in the module's `log` folder, keeping only the current day. Attaching a log when reporting an issue helps considerably.

## How do I report an issue?

Please open one on [GitHub Issues](https://github.com/Xocio/CZero/issues), or join the [group chat](https://t.me/+lwNKCHw_NktjODRh).

For detailed descriptions of each feature see [Features](/en/guide/features); for field-by-field configuration see [Configuration](/en/guide/configuration).
