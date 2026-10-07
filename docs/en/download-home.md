---
layout: page
title: Download
description: Download CZeroX
navbar: false
footer: false
---

<script setup>
import rel from '../.vitepress/release.json'
</script>

<style>
/* No navbar on this page; on narrow screens VitePress would show an empty "back to top"
   button in the top-left (there is no outline). Hide it in favour of a real home link. */
.VPLocalNav { display: none !important; }
</style>

<a class="dl-home" href="/en/">
  <img src="/logo.png" alt="" />
  <span>CZero</span>
</a>

<div class="dl-wrap">
  <header class="dl-head">
    <img class="dl-logo" src="/logo.png" alt="CZero" />
    <h1 class="dl-title">Download CZeroX</h1>
    <p class="dl-sub">{{ rel.version }}</p>
  </header>
  <DownloadEditions />
  <p class="dl-foot"><a href="/en/guide/getting-started">Install guide</a><span class="dl-dot">·</span><a href="/en/download">All files</a></p>
</div>
