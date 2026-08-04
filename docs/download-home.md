---
layout: page
title: 下载
description: 下载 CZero 模块与 CZeroX 应用
navbar: false
footer: false
---

<script setup>
import { ref } from 'vue'
import rel from './.vitepress/release.json'

const got = ref({ mod: false })

const MIRROR = 'https://dl.czeropage.top/'
const moduleUrl = MIRROR + rel.module.file
const appUrl = MIRROR + rel.app.file
</script>

<style>
/* 本页没有导航栏，窄屏下 VitePress 默认会在左上角露出一个空的"回到顶部"按钮（因为本页没有大纲）；
   隐藏它，换成一个真正有用的返回首页按钮 */
.VPLocalNav { display: none !important; }

.dl-home {
  position: fixed;
  top: 16px;
  left: 24px;
  z-index: 10;
  display: flex;
  align-items: center;
  font-size: 16px;
  font-weight: 600;
  color: var(--vp-c-text-1);
  transition: opacity .25s;
}
.dl-home:hover { opacity: .8; }
.dl-home img { height: 24px; margin-right: 8px; }
</style>

<a class="dl-home" href="/">
  <img src="/logo.png" alt="" />
  <span>CZero</span>
</a>

<div class="dl-wrap">
  <header class="dl-head">
    <img class="dl-logo" src="/logo.png" alt="CZero" />
    <h1 class="dl-title">CZero</h1>
    <p class="dl-sub">Android Root 清理模块<span class="dl-ver">{{ rel.version }}</span></p>
  </header>
  <p class="dl-note">配套的 CZeroX 应用已内置在模块里，刷入后自动安装，无需单独下载</p>
  <a class="dl-card dl-card-main" :class="{ 'is-done': got.mod }" :href="moduleUrl" @click="got.mod = true">
    <div class="dl-card-head">
      <span class="dl-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><path d="M3.27 6.96 12 12.01l8.73-5.05"/><path d="M12 22.08V12"/></svg>
      </span>
      <span class="dl-badge">刷入</span>
    </div>
    <h2 class="dl-name">CZero 模块</h2>
    <p class="dl-file">{{ rel.module.file }}</p>
    <span class="dl-btn dl-btn-brand">{{ got.mod ? '已下载 · 重新下载' : '下载模块' }}</span>
  </a>
  <p class="dl-fallback">刷入后没有自动装上 CZeroX？<a :href="appUrl">单独下载应用</a></p>
  <p class="dl-foot">刷入模块后需重启<span class="dl-dot">·</span><a href="/guide/getting-started">安装说明</a><span class="dl-dot">·</span><a href="/download">校验与更多信息</a></p>
</div>
