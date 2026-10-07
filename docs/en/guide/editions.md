# Choosing an Edition

CZeroX comes in two editions for two ways of obtaining privileges. Cleaning rules, the scheduling model and most features are the same; the differences come from **how system privileges are obtained** and the feature set that allows.

| | **Root edition** | **Shizuku edition** |
|---|---|---|
| Authorization | Magisk / KernelSU / APatch | Shizuku |
| Module required | Yes: the module handles scheduling and cleaning | No: the cleaning programs and rules ship inside the app |
| Installation | Flash the module zip, then install CZeroX | Install CZeroX only, start Shizuku and grant access |
| Minimum system | Android 9 | Whatever Shizuku requires |

## Feature differences

| Feature | Root edition | Shizuku edition |
|---|---|---|
| WeChat / QQ / Douyin cleaning, other-apps cleaning | Yes | Yes |
| Custom rules, whitelist, rule sources, per-app cache cleaning | Yes | Yes |
| Background suppression, empty-folder cleanup, recycle bin | Yes | Yes |
| File sorting | Yes, free | Yes, BETA, supporter-unlocked |
| Zero assistant, junk scan, cleaning trends | Yes | Yes |
| F2FS garbage collection | Module-provided method with adjustable conditions | Uses the system's idle maintenance; no condition settings |
| fstrim | Yes | Yes |
| Fine-grained suppression | Yes (needs LSPosed) | Not available |
| Storage tab (device lifespan, live throughput, benchmark) | Yes | Not available |
| Cleaning apps' internal data directories | Yes | Depends on how Shizuku is started; see below |

::: tip Why the differences
Shizuku grants restricted system privileges. Features tied to low-level storage nodes (F2FS status, device lifespan, benchmark) and fine-grained suppression, which relies on the system framework, cannot work that way. This follows from the privilege model and is not a missing feature.
:::

## Which to choose

- Rooted device: choose the **Root edition**. It has the most complete feature set and no dependency on the Shizuku service.
- Non-rooted device: choose the **Shizuku edition**.

## Notes for the Shizuku edition

- **Shizuku must stay running.** The app itself holds no system privileges and everything goes through Shizuku. If Shizuku is not running or not authorized the app says so, and scheduled cleaning does not run.
- **Behavior after a reboot depends on how Shizuku was started.** Shizuku started via Root is available again after boot and CZeroX resumes its scheduled tasks. Shizuku started via ADB stops on every reboot and must be started once more before scheduled tasks resume.
- **Internal data directories cannot be cleaned when started via ADB.** ADB grants fewer privileges than Root, so rule paths pointing into apps' internal data directories have no effect; other paths are unaffected. The status page shows the current privilege level.
- **Status and maintenance.** The status page shows whether the Shizuku service, authorization and the cleaning daemon are healthy, with "Request authorization" and "Redeploy" actions. If scheduled tasks stop unexpectedly, redeploying restores them.
- **Uninstalling.** After uninstalling, revoke the app's authorization from Shizuku's authorized list.

Everything else works as in the Root edition; see [Features](/en/guide/features) and [Configuration](/en/guide/configuration).
