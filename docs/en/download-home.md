---
layout: page
title: Download
description: Download the CZero module and the CZeroX app
navbar: false
footer: false
---

<script setup>
import { ref } from 'vue'
import rel from '../.vitepress/release.json'

const got = ref({ mod: false })

const MIRROR = 'https://dl.czeropage.top/'
const moduleUrl = MIRROR + rel.module.file
const appUrl = MIRROR + rel.app.file
</script>

<style>
/* This page has no navbar, so on narrow screens VitePress falls back to an empty
   "back to top" button top-left (no outline on this page). Hide it and swap in a
   real "back to home" link instead. */
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

<a class="dl-home" href="/en/">
  <img src="/logo.png" alt="" />
  <span>CZero</span>
</a>

<div class="dl-wrap">
  <header class="dl-head">
    <img class="dl-logo" src="/logo.png" alt="CZero" />
    <h1 class="dl-title">CZero</h1>
    <p class="dl-sub">Android Root Cleaning Module<span class="dl-ver">{{ rel.version }}</span></p>
  </header>
  <p class="dl-note">The companion CZeroX app now ships inside the module and installs automatically when you flash it — no separate download needed</p>
  <a class="dl-card dl-card-main" :class="{ 'is-done': got.mod }" :href="moduleUrl" @click="got.mod = true">
    <div class="dl-card-head">
      <span class="dl-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><path d="M3.27 6.96 12 12.01l8.73-5.05"/><path d="M12 22.08V12"/></svg>
      </span>
      <span class="dl-badge">Flash</span>
    </div>
    <h2 class="dl-name">CZero Module</h2>
    <p class="dl-file">{{ rel.module.file }}</p>
    <span class="dl-btn dl-btn-brand">{{ got.mod ? 'Downloaded · Again' : 'Download module' }}</span>
  </a>
  <p class="dl-fallback">Didn't get CZeroX installed automatically? <a :href="appUrl">Download the app separately</a></p>
  <p class="dl-foot">Reboot after flashing<span class="dl-dot">·</span><a href="/en/guide/getting-started">Install guide</a><span class="dl-dot">·</span><a href="/en/download">Checksums &amp; details</a></p>
</div>
