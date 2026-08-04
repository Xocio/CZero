

# 第三方规则源

这个目录收录社区提交的规则源。每个文件是一份独立的清理规则清单，用户在 CZeroX 里填链接即可订阅。

完整的格式说明与字段表见文档站：[开发规则源](https://czeropage.top/guide/rule-source)。

## 怎么投稿

1. 按格式写好 JSON，**先在自己设备上实测一轮**，确认清理量和应用行为都正常
2. 开一个 [issue](https://github.com/Xocio/CZero/issues)，标题注明是规则源投稿，正文附上完整 JSON（或指向你自己仓库的链接）
3. 维护者审核，通过后上传到这里

想修改已收录的规则源，同样开 issue 说明改动。

## 命名

```
<作者>-<主题>.json      例如 laowang-social.json
```

## 收录后的订阅地址

```
https://raw.githubusercontent.com/Xocio/CZero/main/rules/public/<文件名>.json
```

## 最低要求

- 顶层必须有 `name`；建议一并写 `author`、`version`
- 路径必须是绝对路径，不能含 `..`
- 单文件不超过 2 MB、5000 条路径
- 按应用分组，一个应用一组，组名用应用名
- **只清可再生的数据**：缓存、日志、崩溃转储、临时下载。聊天记录、草稿、下载目录、账号凭据一律不要碰
- 拿不准的分组写 `"enabled": false`，让用户自己决定要不要开

审核主要看安全性——有没有指向用户数据的路径、有没有过宽的通配、目录是不是真的只装缓存。

## 示例

```json
{
  "name": "老王的规则源",
  "author": "@laowang",
  "version": "2026.08.04",
  "homepage": "https://github.com/laowang/czero-rules",
  "groups": [
    {
      "name": "微博",
      "enabled": true,
      "paths": [
        "/storage/emulated/0/Android/data/com.sina.weibo/cache",
        "/data/data/com.sina.weibo/files/weibo_log"
      ]
    }
  ]
}
```
