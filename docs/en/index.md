---
# https://vitepress.dev/reference/default-theme-home-page
layout: home

hero:
  name: "CZero"
  text: "System-level Cleanup for Android"
  tagline: Cache cleaning for frequently used apps, plus background suppression, empty-folder cleanup, F2FS garbage collection and fstrim. No resident service, changes take effect instantly.
  image:
    src: /logo.png
    alt: CZero
  actions:
    - theme: brand
      text: Download
      link: /en/download
    - theme: alt
      text: Get Started
      link: /en/guide/getting-started
    - theme: alt
      text: GitHub
      link: https://github.com/Xocio/CZero

features:
  - title: Cache Cleaning
    details: Per-app cleaning scripts for frequently used apps, triggered on schedule and gated by whether the app is actually running, with a minimum interval between cleans of the same app.
  - title: Background Suppression
    details: Periodically detects and suppresses target apps running in the background, cutting needless memory and battery use.
  - title: F2FS GC & fstrim
    details: Reclaims segments when the free ratio gets tight, only with the screen off and battery in good shape — plus a separate fstrim job to keep write performance healthy.
  - title: File Sorting
    details: Sorts files from download folders and similar locations into category folders by type, fully reversible — plus a junk scan and the Zero assistant.
  - title: No Resident Service
    details: A lightweight C++ daemon schedules every task according to config.json, at near-zero cost.
  - title: Hot-Reload Config
    details: The daemon watches config.json and applies changes instantly; a broken config never interrupts the last-good job set.
  - title: Native Companion App
    details: All configuration happens through the native CZeroX app, with live status and stats — no WebUI required.
---
