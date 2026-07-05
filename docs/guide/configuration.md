# 配置参考

CZero 的全部行为集中在一份配置文件：

```
/data/adb/modules/CZero/config.json
```

它是唯一的配置源，由守护进程与每个清理组件直接读取。推荐使用 [CZeroX](/guide/app) 编辑；手动修改时请用**原子写入**（先写临时文件再覆盖），避免守护进程读到半截内容。

::: tip 即时生效
修改保存后**无需重启**。守护进程监视文件变更，会在下一分钟热重载调度；启用项、阈值等非调度字段由各清理组件运行时自读，即时生效。
:::

## 完整示例

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
    "wechat": { "enabled": true, "enhanced": false },
    "qq":     { "enabled": true, "enhanced": false },
    "douyin": { "enabled": true, "enhanced": false },
    "other":  { "enabled": true, "schedule": { "every": "P1D", "at": "03:00" } }
  },
  "suppress": {
    "enabled": true,
    "detect_schedule": { "every": "PT1M" }
  },
  "gc": {
    "enabled": true,
    "dirty_threshold": 200,
    "clean_threshold": 100,
    "schedule": { "every": "PT4H" },
    "script": "/data/adb/modules/CZero/list/GCclean/GCclean1",
    "wait_screen_off_timeout": 30,
    "max_runtime_sec": 600
  },
  "empty_folder": {
    "enabled": true,
    "schedule": { "every": "P1D", "at": "04:00" }
  }
}
```

布尔值使用 JSON 原生 `true` / `false`。

## 字段说明

### general

| 字段 | 类型 | 说明 |
|---|---|---|
| `auto_clean` | bool | 自动缓存清理总开关 |
| `log` | bool | 统一日志开关 |
| `notification` | bool | 清理完成通知 |
| `temporal_barrier_days` | int | 时序屏障：只清理 N 天前的文件，`0` = 不启用 |

### app_clean

| 字段 | 类型 | 说明 |
|---|---|---|
| `detect_schedule` | schedule | 检测应用前台状态的频率（仅分/时） |
| `wechat` / `qq` / `douyin` | object | 各应用的 `enabled` 与 `enhanced`（增强模式）开关 |
| `other.enabled` | bool | 其他应用清理开关 |
| `other.schedule` | schedule | 其他应用清理的执行计划（可多天 + 时刻） |

### suppress

| 字段 | 类型 | 说明 |
|---|---|---|
| `enabled` | bool | 后台压制开关 |
| `detect_schedule` | schedule | 压制检测频率（仅分/时） |

### gc

| 字段 | 类型 | 说明 |
|---|---|---|
| `enabled` | bool | GC 开关 |
| `dirty_threshold` | int | 脏段大于此值触发 GC（须大于 `clean_threshold`） |
| `clean_threshold` | int | 脏段小于此值视为完成 |
| `schedule` | schedule | GC 检测频率（仅分/时） |
| `script` | string | GC 清理脚本路径 |
| `wait_screen_off_timeout` | int | 等待息屏超时（秒） |
| `max_runtime_sec` | int | GC 最大运行时长（秒） |

### empty_folder

| 字段 | 类型 | 说明 |
|---|---|---|
| `enabled` | bool | 空文件夹清理开关 |
| `schedule` | schedule | 执行计划（可多天 + 时刻） |

## schedule 对象

所有 `schedule` / `detect_schedule` 都采用同一种写法：

```json
{ "every": "<ISO8601 时长>", "at": "<HH:MM>" }
```

`every` 表示"每隔多久"，`at` 为可选的执行时刻（仅周期 ≥ 1 天时有意义）。详见 [调度模型](/guide/schedule)。
