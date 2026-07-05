# Install & Setup

## Requirements

- Android 9+ (API 28), `arm64-v8a`
- Root via Magisk, KernelSU, or APatch
- F2FS `/data` partition (only for the GC feature; everything else works regardless)

## Installation

1. Download the latest module zip from [Releases](https://github.com/Xocio/CZero/releases).
2. Flash the zip in Magisk / KernelSU / APatch.
3. Follow the volume-key prompts during flashing:

   | Prompt | Action |
   |---|---|
   | Choose language | Volume Up = English / Volume Down = 中文 |
   | Inherit old config | Volume keys to choose Y / N |

   > Inheritance only carries over the black/white lists and custom path list; `config.json` is always reset to the new default.

4. Reboot.
5. Install the **CZeroX** app to view status and adjust the configuration.

::: tip First install
Not sure what to pick? Just keep the defaults — they already enable cache cleaning, background suppression, GC, and empty-folder cleanup out of the box.
:::

## Verify it's working

After rebooting, confirm the module is running by either:

- Checking the CZero module description in your root manager — it should show an active state.
- Opening **CZeroX** — the home page shows the daemon status (PID / memory) and cleaning stats.

## Updating

Just flash the new zip over the old install. If you choose **inherit config** during flashing, your black/white lists and custom path list are kept; but `config.json` always resets to the new default — re-adjust it (or reset via CZeroX) after updating if needed.

## Uninstalling

Remove the CZero module in your root manager and reboot. The module directory `/data/adb/modules/CZero/` is cleared, including its stats and logs.

## Install troubleshooting

- **Volume keys unresponsive / no key detected** — on some devices `getevent` is slow; after ~60s with no input the installer defaults to Chinese and continues, so it won't hang.
- **No effect after flashing** — make sure you rebooted; then check the daemon status on the CZeroX home page and restart the process in one tap if needed.
- **GC not working** — confirm `/data` is F2FS and `gc.enabled` is `true`. Disable GC on non-F2FS devices.

More in the [FAQ](/en/guide/faq).

## Next steps

- [How It Works](/en/guide/how-it-works) — the full path from boot to cleaning.
- [Features](/en/guide/features) — what each feature actually does.
- [Configuration](/en/guide/configuration) — every `config.json` field explained.
- [Scheduling](/en/guide/schedule) — how `every` / `at` work.
