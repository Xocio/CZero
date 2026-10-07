# 配置参考

::: info Shizuku 版
本页以 Root 版为准。Shizuku 版没有模块目录，配置由应用自行部署并管理，建议仅通过 CZeroX 修改；其 `gc` 一节没有触发条件类字段，也没有 `freezer` 一节。详见[版本选择](/guide/editions)。
:::

CZero 的全部行为集中在一份配置文件：

```
/data/adb/modules/CZero/config.json
```

它是唯一的配置源，由守护进程与每个清理组件直接读取。推荐使用 [CZeroX](/guide/app) 编辑；手动修改时请用**原子写入**（先写临时文件再覆盖），避免守护进程读到半截内容。

::: tip 即时生效
修改保存后**无需重启**。守护进程监视文件变更，会在下一分钟热重载调度；启用项、阈值等非调度字段由各清理组件运行时自读，即时生效。
:::

## 完整示例

下面就是模块自带的默认 `config.json`：

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
    "min_interval_hours": 12,
    "wechat": { "enabled": true },
    "qq":     { "enabled": true },
    "douyin": { "enabled": true },
    "other":  { "enabled": true, "schedule": { "every": "P1D", "at": "03:00" } }
  },
  "suppress": {
    "enabled": true,
    "detect_schedule": { "every": "PT1M" }
  },
  "freezer": {
    "auto_enable": false,
    "disable_vendor_guard": false
  },
  "gc": {
    "enabled": true,
    "free_percent": 95,
    "target_percent": 98,
    "min_interval_hours": 6,
    "require_charging": false,
    "min_battery": 40,
    "max_battery_temp": 400,
    "schedule": { "every": "PT12H" },
    "script": "/data/adb/modules/CZero/list/GCclean/GCclean1",
    "wait_screen_off_timeout": 30,
    "max_runtime_sec": 420
  },
  "trim": {
    "enabled": true,
    "schedule": { "every": "P7D", "at": "04:00" },
    "script": "/data/adb/modules/CZero/list/GCclean/GCclean1 trim"
  },
  "empty_folder": {
    "enabled": true,
    "schedule": { "every": "P1D", "at": "04:00" }
  },
  "file_sort": {
    "enabled": false,
    "root": "/storage/emulated/0/CZero",
    "quiet_sec": 60,
    "depth": 3,
    "duplicates": true,
    "schedule": { "every": "P1D", "at": "05:00" },
    "folders": {
      "package": "Apps",
      "archive": "Archives",
      "document": "Documents",
      "image": "Images",
      "video": "Videos",
      "audio": "Audio",
      "code": "Code",
      "font": "Fonts",
      "other": "Others",
      "duplicate": "Duplicates"
    }
  }
}
```

布尔值使用 JSON 原生 `true` / `false`。

::: tip 缺字段不会出事
每个组件都自带一份默认值，读不到或解析失败的单个字段会**回落到默认值**，不影响其他字段。所以你不必把上面所有键都写全。
:::

## 字段说明

### general

| 字段 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `auto_clean` | bool | `true` | 自动缓存清理总开关 |
| `log` | bool | `false` | 统一日志开关 |
| `notification` | bool | `false` | 清理完成通知（含超级岛） |
| `temporal_barrier_days` | int | `3` | 时序屏障：只清理 N 天前的文件，`0` = 不启用 |
| `recycle_enabled` | bool | `true` | **可选**。回收站开关；关闭后清理将直接彻底删除，无法恢复 |
| `recycle_keep_days` | int | `7` | **可选**。回收站保留天数，`0` 或缺省即 7 天；仅影响之后新产生的记录 |

### app_clean

| 字段 | 类型 | 说明 |
|---|---|---|
| `detect_schedule` | schedule | 检测应用前台状态的频率（仅分/时） |
| `min_interval_hours` | int | 同一应用两次自动清理的最小间隔（小时），默认 `12` |
| `wechat` / `qq` / `douyin` | object | 各应用的 `enabled` 开关 |
| `other.enabled` | bool | 其他应用清理开关 |
| `other.schedule` | schedule | 其他应用清理的执行计划（可多天 + 时刻） |

### suppress

| 字段 | 类型 | 说明 |
|---|---|---|
| `enabled` | bool | 后台压制开关 |
| `detect_schedule` | schedule | 压制检测频率（仅分/时） |

压制的目标应用与进程名单不在 `config.json` 里，由 CZeroX 单独管理。详见[功能详解 · 后台压制](/guide/features#后台压制)。

### freezer

**精细化压制**的开机动作。它压制后台应用而不杀死，应用状态完整保留、切回瞬间恢复，详见[功能详解 · 精细化压制](/guide/features#精细化压制)。

| 字段 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `auto_enable` | bool | `false` | 每次开机打开系统的精细化压制能力（需要 Android 12+） |
| `disable_vendor_guard` | bool | `false` | 关掉厂商自研的保活/压制，避免和精细化压制互相打架 |

::: warning disable_vendor_guard 有前置条件
只有在 CZeroX 的精细化压制引擎已启用时它才会动手，否则直接跳过，不会平白改动厂商设置。
:::

参与压制的应用名单不在 `config.json` 里，由 CZeroX 单独管理。

### gc

F2FS 垃圾回收。触发看的是**空闲比例**，不是绝对的脏段数。

| 字段 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `enabled` | bool | `true` | GC 开关 |
| `free_percent` | int | `95` | 空闲比例**低于**此值才触发回收；上限 99 |
| `target_percent` | int | `98` | 空闲比例回到此值即视为完成；上限 100，且必须大于 `free_percent`，否则自动抬高 1 个百分点 |
| `min_interval_hours` | int | `6` | 两次回收的最小间隔（小时），最小 1 |
| `require_charging` | bool | `false` | 为 `true` 时只在充电中执行 |
| `min_battery` | int | `40` | 非充电时的最低电量（%），低于则跳过 |
| `max_battery_temp` | int | `400` | 电池温度上限，单位 **0.1°C**（`400` = 40.0°C）；最小 100 |
| `schedule` | schedule | `PT12H` | 多久检查一次（仅分/时） |
| `script` | string | — | GC 程序路径，一般不需要改 |
| `wait_screen_off_timeout` | int | `30` | 等待息屏超时（秒），最小 5；超时仍亮屏则本轮放弃 |
| `max_runtime_sec` | int | `420` | 单次 GC 最长运行时长（秒），最小 60 |

::: tip schedule 调密没用
`schedule` 只决定"多久去看一眼"，真正跑不跑还要过 `min_interval_hours` 那一关。
:::

### trim

fstrim，与 GC **独立调度**、互不影响。

| 字段 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `enabled` | bool | `true` | fstrim 开关 |
| `schedule` | schedule | `P7D` + `04:00` | 执行计划（可多天 + 时刻） |
| `script` | string | — | 程序路径与参数，一般不需要改 |

### empty_folder

| 字段 | 类型 | 说明 |
|---|---|---|
| `enabled` | bool | 空文件夹清理开关 |
| `schedule` | schedule | 执行计划（可多天 + 时刻） |

### file_sort

文件归类（Root 版已正式提供，Shizuku 版为 BETA），详见[功能详解 · 文件归类](/guide/features#文件归类)。

| 字段 | 类型 | 默认 | 说明 |
|---|---|---|---|
| `enabled` | bool | `false` | 自动归类开关；关闭时仍可在 CZeroX 中手动归类 |
| `root` | string | `/storage/emulated/0/CZero` | 归类目录，需为以 `/` 开头的完整路径 |
| `quiet_sec` | int | `60` | 文件写入完成后静置多久再移动（秒），范围 0–3600 |
| `depth` | int | `3` | 来源目录向下扫描的层数，`0` = 不限层级 |
| `duplicates` | bool | `true` | 同名且同大小的文件是否移入重复目录 |
| `schedule` | schedule | `P1D` + `05:00` | 兜底执行计划；平时由文件写入触发，该计划用于补漏 |
| `folders` | object | 见上方示例 | 各分类的目录名，不能包含 `/` |

来源目录与各分类的启停在 CZeroX 中管理。

## schedule 对象

所有 `schedule` / `detect_schedule` 都采用同一种写法：

```json
{ "every": "<ISO8601 时长>", "at": "<HH:MM>" }
```

`every` 表示"每隔多久"，`at` 为可选的执行时刻（仅周期 ≥ 1 天时有意义）。详见 [调度模型](/guide/schedule)。

## 不在 config.json 里的配置

有几类数据体量大、结构不同，不放在 `config.json` 里，统一在 CZeroX 中管理：

- 自定义清理路径与清理白名单（含订阅的[规则源](/guide/rule-source)）
- 按应用组织的缓存规则
- 后台压制与精细化压制的应用名单
- 空文件夹的清扫范围与白名单
- 文件归类的来源目录

重装模块时选择继承，这些都会被沿用。
