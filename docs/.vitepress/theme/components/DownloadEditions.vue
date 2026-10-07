<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useData } from 'vitepress'
import rel from '../../release.json'

/**
 * 下载区
 */

type Key = 'root' | 'shizuku'

const MIRROR = 'https://dl.czeropage.top/'
const release = rel as Record<string, any>
const urlOf = (a?: { file?: string }) => (a && a.file ? MIRROR + a.file : '')

/** inline：放在文档正文里时使用，带浅底并在顶部显示版本号 */
const props = defineProps<{ inline?: boolean }>()

const { lang } = useData()
const en = computed(() => lang.value.startsWith('en'))

const KEYS: Key[] = ['root', 'shizuku']
const current = ref<Key>('root')

onMounted(() => {
  if (location.hash === '#shizuku') current.value = 'shizuku'
})

function select(key: Key) {
  current.value = key
  history.replaceState(null, '', key === 'shizuku' ? '#shizuku' : location.pathname + location.search)
}

const editions = computed(() => {
  const e = en.value
  return {
    root: {
      tab: 'Root',
      desc: e ? 'For Magisk, KernelSU and APatch' : '适用于 Magisk、KernelSU 与 APatch',
      btn: e ? 'Download module' : '下载模块',
      url: urlOf(release.module),
    },
    shizuku: {
      tab: 'Shizuku',
      desc: e ? 'No root required. Works with Shizuku' : '无需 Root，配合 Shizuku 使用',
      btn: e ? 'Download app' : '下载应用',
      url: urlOf(release.shizuku),
    },
  }
})

const ed = computed(() => editions.value[current.value])
const unavailable = computed(() => (en.value ? 'Not available yet' : '暂未提供'))
</script>

<template>
  <div class="dl vp-raw" :class="{ 'is-inline': props.inline }">
    <p v-if="props.inline" class="dl-ver-line">{{ release.version }}</p>

    <div class="dl-seg" role="tablist" :aria-label="en ? 'Edition' : '版本'">
      <button
        v-for="key in KEYS"
        :key="key"
        type="button"
        role="tab"
        class="dl-seg-item"
        :class="{ active: current === key }"
        :aria-selected="current === key"
        @click="select(key)"
      >
        {{ editions[key].tab }}
      </button>
    </div>

    <p class="dl-desc" role="tabpanel">{{ ed.desc }}</p>

    <a v-if="ed.url" class="dl-btn" :href="ed.url">{{ ed.btn }}</a>
    <span v-else class="dl-btn is-off" aria-disabled="true">{{ unavailable }}</span>
  </div>
</template>

<style scoped>
.dl {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.dl-ver-line {
  margin: 0 0 20px;
  font-size: 13px;
  line-height: 1.5;
  color: var(--vp-c-text-3);
}

/* 分段切换：与 VitePress 自身的圆角按钮语汇保持一致 */
.dl-seg {
  display: inline-flex;
  padding: 3px;
  border-radius: 999px;
  background: var(--vp-c-default-soft);
}

.dl-seg-item {
  min-width: 104px;
  padding: 6px 18px;
  border-radius: 999px;
  font-size: 14px;
  font-weight: 500;
  line-height: 1.5;
  color: var(--vp-c-text-2);
  transition: color .2s, background-color .2s, box-shadow .2s;
}

.dl-seg-item:hover {
  color: var(--vp-c-text-1);
}

.dl-seg-item.active {
  color: var(--vp-c-text-1);
  background: var(--vp-c-bg);
  box-shadow: 0 1px 2px rgba(0, 0, 0, .08);
}

.dark .dl-seg-item.active {
  background: var(--vp-c-bg-elv);
  box-shadow: none;
}

.dl-seg-item:focus-visible,
.dl-btn:focus-visible {
  outline: 2px solid var(--vp-c-brand-1);
  outline-offset: 2px;
}

.dl-desc {
  margin: 24px 0 0;
  font-size: 15px;
  line-height: 1.6;
  color: var(--vp-c-text-2);
}

.dl-btn {
  display: inline-block;
  min-width: 200px;
  margin-top: 28px;
  padding: 0 32px;
  border-radius: 999px;
  font-size: 15px;
  font-weight: 600;
  line-height: 46px;
  color: var(--vp-button-brand-text);
  background: var(--vp-button-brand-bg);
  text-decoration: none;
  transition: background-color .2s;
}

.dl-btn:hover {
  color: var(--vp-button-brand-hover-text);
  background: var(--vp-button-brand-hover-bg);
}

.dl-btn.is-off {
  color: var(--vp-c-text-3);
  background: var(--vp-c-default-soft);
  cursor: not-allowed;
}

/* 文档正文里：加一层浅底，和上下文分开 */
.dl.is-inline {
  margin: 20px 0 8px;
  padding: 36px 24px 40px;
  border-radius: 12px;
  background: var(--vp-c-bg-soft);
}

@media (max-width: 640px) {
  .dl-desc {
    font-size: 14.5px;
  }

  .dl-btn {
    width: 100%;
  }

  .dl.is-inline {
    padding: 28px 16px 28px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .dl-seg-item,
  .dl-btn {
    transition: none;
  }
}
</style>
