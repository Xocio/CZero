# Authoring Rule Sources

A rule source lets anyone maintain and distribute their own cleaning rules. It is just a JSON file served over HTTP: users paste the link into CZeroX to subscribe, and it **updates automatically once a day**.

::: tip Who this is for
This page is for developers publishing rules to CZero users. **Hosting from your own repo is recommended**; the CZero repo keeps a fallback slot as well.

If you only want a few paths for your own device, add them directly under **Custom rules → Built-in rules** in CZeroX — you don't need a rule source.

The Chinese page is the canonical one and may be more current: [开发规则源](/guide/rule-source).
:::

## File format

A rule source has the same shape as the module's local `clean_paths.json`, with attribution fields added at the top level:

```json
{
  "name": "Wang's rule source",
  "author": "@laowang",
  "version": "2026.08.04",
  "homepage": "https://github.com/laowang/czero-rules",
  "groups": [
    {
      "name": "Weibo",
      "enabled": true,
      "paths": [
        "/storage/emulated/0/Android/data/com.sina.weibo/cache",
        "/data/data/com.sina.weibo/files/weibo_log"
      ]
    },
    {
      "name": "Xiaohongshu",
      "paths": [
        "/storage/emulated/0/Android/data/com.xingin.xhs/cache/*"
      ]
    }
  ]
}
```

Export the rules you curated in CZeroX, add `name` and `author`, and you have a working source.

## Top-level fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | **yes** | Source name, shown as the card title. Truncated past 64 chars |
| `author` | string | no | Attribution, shown under the title. Max 64 chars |
| `version` | string | no | Version marker, any format; a date reads best. Max 32 chars |
| `homepage` | string | no | Project page. Max 256 chars |
| `groups` | array | **yes** | Rule groups; must yield at least one valid path |

A missing `name`, or no valid path anywhere in `groups`, gets the source rejected on add.

## Group fields

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | no | Group name. Empty shows as "Ungrouped" |
| `enabled` | bool | no | **Defaults to `true`**. Set `false` to ship the group off by default; users can turn it on |
| `paths` | array | **yes** | Cleaning paths in this group |

Groups are an organizational device — cleaning flattens them. Still, group carefully: users browse and toggle rules per group, so a source that dumps hundreds of paths into one group is painful to use.

## Writing paths

Paths **must be absolute**. Glob patterns are supported:

| Pattern | Meaning |
|---|---|
| `*` | Any characters within one segment |
| `?` | A single character |
| `[abc]` | One character from the set |
| `**` | Recursive match across directory levels |

Common roots:

```
/storage/emulated/0/Android/data/<package>/   App external private dir
/data/data/<package>/                         App internal data dir
/data/user/0/<package>/                       Same, multi-user form
/data/media/0/                                Equivalent to /storage/emulated/0
```

::: warning Directories are cleaned recursively
A directory path means **every file under it** gets cleaned, then the empty directories are removed. Target real cache dirs like `cache` or `log` — never a parent like `files` that may hold user data.
:::

## Paths that get rejected

Every path is validated on fetch. Offending entries are **dropped individually**; the source itself still loads:

- Relative paths not starting with `/`
- Anything containing `..` (prevents escaping the intended target)
- Longer than 512 chars, or shorter than 2

## Size limits

| Limit | Value | On exceeding |
|---|---|---|
| File size | 2 MB | Whole source rejected |
| Path count | 5000 | Extras dropped; the first 5000 still apply |

## Hosting

### Recommended: your own repo

Any HTTPS endpoint returning JSON works. The easiest route is a GitHub repo of your own:

1. Create a public repo, e.g. `your-name/czero-rules`
2. Put the rule file in it, e.g. `rules.json`
3. Distribute the raw link:

```
https://raw.githubusercontent.com/<user>/<repo>/main/rules.json
```

The response must be bare JSON. Note that a GitHub **web** URL (`github.com/.../blob/...`) returns an HTML page, not JSON — users pasting it will get "The file is not valid JSON".

::: warning http/https only
Schemes like `file://` and `content://` are rejected.
:::

Self-hosting keeps you in full control: what changes and when is up to you. Keeping rules in Git also means you can roll back a bad change and answer user issues against a diff. Bump `version` in the JSON when you publish — users see it change in the app.

::: tip Request volume is not a concern
`raw.githubusercontent.com` rate-limits **per visitor IP**, not per repository, so it doesn't matter how many people subscribe to yours. And each user fetches each source once a day.
:::

### Fallback: the CZero repo

If maintaining a repo really isn't practical for you, open an [issue](https://github.com/Xocio/CZero/issues) asking to place your JSON under `rules/public/` in the CZero repo:

```
https://raw.githubusercontent.com/Xocio/CZero/main/rules/public/<filename>.json
```

::: warning Storage only
CZero does **not** review rule content. Being hosted here implies no vetting and no endorsement — correctness and safety remain the author's responsibility. This exists for people who can't host their own; if you can, do.
:::

## Subscribing

CZeroX → Settings → Custom rules → `⋯` (top right) → Add rule source → paste the link.

A new card appears titled with your `name` and subtitled with `author`, showing group count, rule count and last update time. Tapping it opens the full rule list.

## Update behavior

- When a user opens the rule sources page and more than **24 hours** have passed since the last sync, every enabled source is fetched automatically
- Users can also tap the refresh icon on a card
- An update **replaces the source's rules wholesale** — your file wins
- Re-adding the same URL counts as an update; it will not create a second card

::: warning Local edits are overwritten
Users can add, edit and delete your rules locally (say, to disable a group temporarily), but those edits are wiped on the next update — ownership is decided by the `source` key in the JSON, and updates only trust the source file. This is by design; changes users want to keep belong in their built-in rules.
:::

## Execution

Your rules are handed to `list/customize` alongside the built-in ones and behave identically:

- **Recycle bin**: matched files are moved to `/storage/emulated/0/Recycle`, not deleted outright. Users can restore them, kept 7 days by default
- **Temporal barrier**: if enabled, only files older than N days are cleaned
- **Whitelist**: the user's `clean_whitelist.json` wins; whitelisted paths are skipped
- **Nested dedup**: if a parent directory is already listed, paths under it are ignored automatically

So duplicates are harmless — writing too broadly is not.

## Recommendations

1. **Verify on your own device first.** Load the rules as built-in rules in CZeroX, run a cycle, confirm both the reclaimed size and the app's behavior before publishing.
2. **Only clean regenerable data.** Caches, logs, crash dumps and temp downloads are safe; chat history, drafts, download folders and credentials are not.
3. **One group per app**, named after the app, so users can toggle what they care about.
4. **Bump `version` on every change.** Users can't see a diff, but they can see the version move.
5. **Use `**` sparingly.** Recursive matching is slow on large trees and easy to over-match.
6. **Ship risky groups disabled.** Set `"enabled": false` when unsure and let the user opt in.
