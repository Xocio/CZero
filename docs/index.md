---
# https://vitepress.dev/reference/default-theme-home-page
layout: home

hero:
  name: "CZero"
  text: "Android Root 清理模块"
  tagline: 为高频应用提供缓存清理，并涵盖后台压制、空文件夹清理、F2FS 垃圾回收与 fstrim。无常驻服务，配置即时生效。
  image:
    src: /logo.png
    alt: CZero
  actions:
    - theme: brand
      text: 立即下载
      link: /download
    - theme: alt
      text: 快速上手
      link: /guide/getting-started
    - theme: alt
      text: GitHub
      link: https://github.com/Xocio/CZero

features:
  - title: 缓存清理
    details: 为高频应用提供各自独立的清理脚本，按计划触发并先检测应用是否在运行，可按需开启增强模式。
  - title: 后台压制
    details: 周期性检测并压制在后台运行的目标应用，减少无谓的内存与耗电占用。
  - title: F2FS GC 与 fstrim
    details: 空闲段占比吃紧时回收，息屏、电量与温度都满足才动手；另有独立的 fstrim 任务维持写入性能。
  - title: 无常驻服务
    details: 由一个轻量 C++ 守护进程按 config.json 调度全部任务，几乎不占资源。
  - title: 配置热重载
    details: 守护进程监视 config.json，保存即生效；配置损坏时保留上一份有效任务，绝不中断。
  - title: 原生配套应用
    details: 全部配置通过原生应用 CZeroX 完成，实时查看状态与统计，无需 WebUI。
---
