---
# https://vitepress.dev/reference/default-theme-home-page
layout: home

hero:
  name: "CZero"
  text: "Android Root 清理模块"
  tagline: 为高频应用提供缓存清理，并涵盖后台压制、空文件夹清理与 F2FS 垃圾回收。无常驻服务，配置即时生效。
  image:
    src: /logo.png
    alt: CZero
  actions:
    - theme: brand
      text: 快速上手
      link: /guide/getting-started
    - theme: alt
      text: 什么是 CZero
      link: /guide/what-is-czero
    - theme: alt
      text: GitHub
      link: https://github.com/Xocio/CZero

features:
  - title: 缓存清理
    details: 为高频应用提供各自独立的清理脚本，按计划触发并先检测应用是否在运行，可按需开启增强模式。
  - title: 后台压制
    details: 周期性检测并压制在后台运行的目标应用，减少无谓的内存与耗电占用。
  - title: F2FS GC
    details: 监控脏段数量，超阈值时执行垃圾回收；等待息屏后再运行并限制最长运行时间。
  - title: 无常驻服务
    details: 由一个轻量 C++ 守护进程按 config.json 调度全部任务，几乎不占资源。
  - title: 配置热重载
    details: 守护进程监视 config.json，保存即生效；配置损坏时保留上一份有效任务，绝不中断。
  - title: 原生配套应用
    details: 全部配置通过原生应用 CZeroX 完成，实时查看状态与统计，无需 WebUI。
---
