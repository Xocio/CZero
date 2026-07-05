# Build from Source

The repo contains only the **open-source source files** — build outputs are not committed; releases are published as a packaged module zip under Releases. To modify and recompile, cross-compile for Android ARM64, then repackage into a flashable zip.

> Most users don't need to build — just download the module zip from [Releases](https://github.com/Xocio/CZero/releases) and flash it. This page is for developers who want to change the source.

## Tools required

- **Android NDK** (with `clang++`) for the C++ code.
- **Go toolchain** (1.21+) for the Go cleaners (cross-compiled to `android/arm64`).

## What needs compiling

### C++ (NDK clang++)

| Source | Output |
|---|---|
| `service.cpp` | `service` |
| `cron/timer_daemon.cpp` | `cron/timer_daemon` |
| `list/Tencent/check.cpp` | `list/Tencent/check` |
| `list/suppress/suppress.cpp` | `list/suppress/suppress` |
| `list/GCclean/GCclean1.cpp` | `list/GCclean/GCclean1` |
| `list/Emptyfolder/emptyfolder.cpp` | `list/Emptyfolder/emptyfolder` |
| `list/zero.cpp` | `list/zero` |

### Go (`GOOS=android GOARCH=arm64`)

| Source | Output |
|---|---|
| `list/Tencent/tencentmm.go` | `list/Tencent/tencentmm` |
| `list/Tencent/mobileqq.go` | `list/Tencent/mobileqq` |
| `list/Tencent/ugcaweme.go` | `list/Tencent/ugcaweme` |
| `list/customize.go` | `list/customize` |

## Compile the C++

Run from the repo root, adjusting the NDK path as needed:

```bash
NDK=/path/to/ndk
CXX="$NDK/toolchains/llvm/prebuilt/linux-x86_64/bin/clang++"
FLAGS="--target=aarch64-linux-android28 -O2 -std=c++17 -static-libstdc++ -I."

$CXX $FLAGS -o service                     service.cpp
$CXX $FLAGS -o cron/timer_daemon           cron/timer_daemon.cpp
$CXX $FLAGS -o list/Tencent/check          list/Tencent/check.cpp
$CXX $FLAGS -o list/suppress/suppress      list/suppress/suppress.cpp
$CXX $FLAGS -o list/GCclean/GCclean1       list/GCclean/GCclean1.cpp
$CXX $FLAGS -o list/Emptyfolder/emptyfolder list/Emptyfolder/emptyfolder.cpp
$CXX $FLAGS -o list/zero                   list/zero.cpp
```

::: warning C++ build constraints
- Must be linked with **`-static-libstdc++`** (the device has no `libc++_shared.so`); **keep PIE, never use `-static`**.
- Target **API 28+** (`posix_spawn` in `timer_daemon` requires it).
- Every compile must add **`-I<repo-root>`** (the `-I.` above) — several files share headers under `common/`.
:::

## Compile the Go

```bash
export GOOS=android GOARCH=arm64 CGO_ENABLED=0
go build -o list/Tencent/tencentmm  list/Tencent/tencentmm.go
go build -o list/Tencent/mobileqq   list/Tencent/mobileqq.go
go build -o list/Tencent/ugcaweme   list/Tencent/ugcaweme.go
go build -o list/customize          list/customize.go
```

## Shared headers

Several C++ files depend on shared headers under `common/`, which is why every compile needs `-I<repo-root>`:

- **`common/json_config.h`** — the config parser, shared by `timer_daemon`, `check`, `suppress`, `GCclean1`, `emptyfolder`.
- **`common/file_lock.h`** — the `basis.prop` file-lock helper, shared by `GCclean1`, `emptyfolder`, `suppress`, `zero`.

## Packaging

Once compiled, package all outputs together with the scripts, `config.json`, `module.prop`, `META-INF/`, `list/`, `common/`, etc., preserving the original directory layout and executable permissions, into a zip ready to flash.
