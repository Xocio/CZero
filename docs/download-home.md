---
layout: page
title: 下载
description: 下载 CZeroX
navbar: false
footer: false
---

<script setup>
import rel from './.vitepress/release.json'
</script>

<style>
/* 本页没有导航栏，窄屏下 VitePress 默认会在左上角露出一个空的"回到顶部"按钮（因为本页没有大纲）；
   隐藏它，换成一个真正有用的返回首页按钮 */
.VPLocalNav { display: none !important; }
</style>

<a class="dl-home" href="/">
  <img src="/logo.png" alt="" />
  <span>CZero</span>
</a>

<div class="dl-wrap">
  <header class="dl-head">
    <img class="dl-logo" src="/logo.png" alt="CZero" />
    <h1 class="dl-title">下载 CZeroX</h1>
    <p class="dl-sub">{{ rel.version }}</p>
  </header>
  <DownloadEditions />
  <p class="dl-foot"><a href="/guide/getting-started">安装说明</a><span class="dl-dot">·</span><a href="/download">全部文件</a></p>
</div>
