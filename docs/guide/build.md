# 从源码构建

仓库只包含**开源的源码文件**，编译产物不入库；发布则以打包好的模块 zip 形式放在 Releases。若你想自行修改并重新编译，需要为 Android ARM64 交叉编译后再打包成刷机 zip。

> 大多数用户无需构建 —— 直接从 [Releases](https://github.com/Xocio/CZero/releases) 下载模块 zip 刷入即可。本页面向希望改动源码的开发者。

## 需要的工具

- **Android NDK**（含 `clang++`），用于编译 C++。
- **Go 工具链**（1.21+），用于编译 Go 清理器（交叉编译到 `android/arm64`）。

## 需要编译的文件

### C++（NDK clang++）

| 源码 | 产物 |
|---|---|
| `service.cpp` | `service` |
| `cron/timer_daemon.cpp` | `cron/timer_daemon` |
| `list/Tencent/check.cpp` | `list/Tencent/check` |
| `list/suppress/suppress.cpp` | `list/suppress/suppress` |
| `list/GCclean/GCclean1.cpp` | `list/GCclean/GCclean1` |
| `list/Emptyfolder/emptyfolder.cpp` | `list/Emptyfolder/emptyfolder` |
| `list/zero.cpp` | `list/zero` |

### Go（`GOOS=android GOARCH=arm64`）

| 源码 | 产物 |
|---|---|
| `list/Tencent/tencentmm.go` | `list/Tencent/tencentmm` |
| `list/Tencent/mobileqq.go` | `list/Tencent/mobileqq` |
| `list/Tencent/ugcaweme.go` | `list/Tencent/ugcaweme` |
| `list/customize.go` | `list/customize` |

## 编译 C++

从仓库根目录运行，按需调整 NDK 路径：

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

::: warning C++ 编译约束
- 必须使用 **`-static-libstdc++`** 链接（设备没有 `libc++_shared.so`）；**保持 PIE，切勿 `-static`**。
- 目标 **API 28+**（`timer_daemon` 使用的 `posix_spawn` 需要）。
- 每次编译都要加 **`-I<仓库根目录>`**（即上面的 `-I.`）—— 多个文件共用 `common/` 下的头文件。
:::

## 编译 Go

```bash
export GOOS=android GOARCH=arm64 CGO_ENABLED=0
go build -o list/Tencent/tencentmm  list/Tencent/tencentmm.go
go build -o list/Tencent/mobileqq   list/Tencent/mobileqq.go
go build -o list/Tencent/ugcaweme   list/Tencent/ugcaweme.go
go build -o list/customize          list/customize.go
```

## 共享头文件

多个 C++ 文件依赖 `common/` 下的公共头，这也是每次编译都必须 `-I<仓库根目录>` 的原因：

- **`common/json_config.h`** —— 配置解析器，被 `timer_daemon`、`check`、`suppress`、`GCclean1`、`emptyfolder` 共用。
- **`common/file_lock.h`** —— `basis.prop` 的文件锁助手，被 `GCclean1`、`emptyfolder`、`suppress`、`zero` 共用。

## 打包

编译完成后，将所有产物连同脚本、`config.json`、`module.prop`、`META-INF/`、`list/`、`common/` 等按原有目录结构打包为 zip 即可用于刷入。注意保留可执行权限。
