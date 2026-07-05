<p align="center">
  <img src="assets/logo.png" width="120" alt="CZero logo">
</p>

<h1 align="center">CZero</h1>

<p align="center">Striving to be the best cache & junk cleaning solution on Android.</p>

<p align="center">
  <a href="https://github.com/Xocio/CZero/releases"><img src="https://img.shields.io/github/v/release/Xocio/CZero?label=release&color=orange" alt="release"></a>
  <a href="https://czeropage.top/en/"><img src="https://img.shields.io/badge/docs-website-blue" alt="docs"></a>
  <img src="https://img.shields.io/badge/Android-9%2B-3DDC84?logo=android&logoColor=white" alt="Android 9+">
  <img src="https://img.shields.io/badge/root-Magisk%20%7C%20KernelSU%20%7C%20APatch-red" alt="root">
  <a href="LICENSE"><img src="https://img.shields.io/github/license/Xocio/CZero?color=blue" alt="license"></a>
</p>

<p align="center"><a href="README.md">简体中文</a> · <b>English</b></p>

---

CZero is an Android root module that cleans the cache of frequently used apps, and adds background suppression, empty-folder cleanup, and F2FS garbage collection.

The module has no resident service — a lightweight C++ daemon schedules every task according to `config.json`, and configuration changes take effect immediately. Day-to-day operation goes through the native companion app **CZeroX**.

## Documentation

> **Official docs [DOCS](https://czeropage.top/en/)**
>
> Reading the docs before use is strongly recommended; if you run into trouble, check the [FAQ](https://czeropage.top/en/guide/faq) first.

## Features

- **Cache cleaning** — per-app cleaning scripts for frequently used apps, triggered on schedule and gated by whether the app is actually running; an optional enhanced mode is available.
- **Background suppression** — periodically detects and suppresses the target apps.
- **F2FS GC** — monitors dirty segments and runs garbage collection past a threshold, waiting for screen-off with a runtime cap.
- **Along the way** — custom-path cleaning (`clean_paths.prop`) and empty-folder sweeping.
- **Hot-reload config** — the daemon watches `config.json` and applies changes instantly; a broken config never kills the last-good job set.
- **Optional stats & logging** — per-app cleaning counts and a tagged daily log, both easy to turn off.

## Usage

1. Download the latest module zip from [Releases](https://github.com/Xocio/CZero/releases).
2. Flash it in Magisk / KernelSU / APatch and use the volume keys to pick a language and whether to inherit your old config.
3. Reboot, install **CZeroX**, and configure to taste.

## CZeroX

The native Jetpack Compose companion app, styled with [Miuix](https://compose-miuix-ui.github.io/miuix/).

<p align="center">
  <img src="assets/webx.png" width="260" alt="CZeroX home">
</p>

## Star History

<a href="https://www.star-history.com/?repos=Xocio%2FCZero&type=timeline&logscale=&legend=top-left">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/chart?repos=Xocio/CZero&type=timeline&theme=dark&logscale&legend=top-left" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/chart?repos=Xocio/CZero&type=timeline&logscale&legend=top-left" />
   <img alt="Star History Chart" src="https://api.star-history.com/chart?repos=Xocio/CZero&type=timeline&logscale&legend=top-left" />
 </picture>
</a>

## License

[Apache License 2.0](LICENSE)
