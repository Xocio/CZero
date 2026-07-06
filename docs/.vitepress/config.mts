import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  base: '/',
  title: 'CZero',
  description: 'Android Root 清理模块 · 缓存清理 / 后台压制 / F2FS GC',
  lang: 'zh-CN',
  cleanUrls: true,
  lastUpdated: true,

  head: [
    ['link', { rel: 'icon', href: '/logo.png' }],
  ],

  themeConfig: {
    logo: '/logo.png',
    socialLinks: [
      { icon: 'github', link: 'https://github.com/Xocio/CZero' },
    ],
    search: {
      provider: 'local',
    },
  },

  locales: {
    root: {
      label: '简体中文',
      lang: 'zh-CN',
      themeConfig: {
        nav: [
          { text: '指南', link: '/guide/what-is-czero', activeMatch: '/guide/' },
          { text: '配置', link: '/guide/configuration' },
          {
            text: 'v1.0',
            items: [
              { text: '下载', link: 'https://github.com/Xocio/CZero/releases' },
              { text: '更新日志', link: 'https://github.com/Xocio/CZero/releases' },
            ],
          },
        ],
        sidebar: {
          '/guide/': [
            {
              text: '入门',
              items: [
                { text: '什么是 CZero', link: '/guide/what-is-czero' },
                { text: '安装与上手', link: '/guide/getting-started' },
                { text: '工作原理', link: '/guide/how-it-works' },
              ],
            },
            {
              text: '使用',
              items: [
                { text: '功能详解', link: '/guide/features' },
                { text: '配置参考', link: '/guide/configuration' },
                { text: '调度模型', link: '/guide/schedule' },
                { text: 'CZeroX 应用', link: '/guide/app' },
              ],
            },
            {
              text: '进阶',
              items: [
                { text: '从源码构建', link: '/guide/build' },
                { text: '常见问题', link: '/guide/faq' },
              ],
            },
          ],
        },
        footer: {
          message: '基于 <a href="https://github.com/Xocio/CZero/blob/main/LICENSE" target="_blank" rel="noopener">GPL-3.0</a> 许可发布',
          copyright: 'Copyright © 2026 Xocio',
        },
        docFooter: { prev: '上一页', next: '下一页' },
        outline: { label: '本页目录' },
        lastUpdated: { text: '最后更新' },
        returnToTopLabel: '回到顶部',
        sidebarMenuLabel: '菜单',
        darkModeSwitchLabel: '主题',
        lightModeSwitchTitle: '切换到浅色模式',
        darkModeSwitchTitle: '切换到深色模式',
      },
    },

    en: {
      label: 'English',
      lang: 'en-US',
      link: '/en/',
      themeConfig: {
        nav: [
          { text: 'Guide', link: '/en/guide/what-is-czero', activeMatch: '/en/guide/' },
          { text: 'Config', link: '/en/guide/configuration' },
          {
            text: 'v1.0',
            items: [
              { text: 'Download', link: 'https://github.com/Xocio/CZero/releases' },
              { text: 'Changelog', link: 'https://github.com/Xocio/CZero/releases' },
            ],
          },
        ],
        sidebar: {
          '/en/guide/': [
            {
              text: 'Getting Started',
              items: [
                { text: 'What is CZero', link: '/en/guide/what-is-czero' },
                { text: 'Install & Setup', link: '/en/guide/getting-started' },
                { text: 'How It Works', link: '/en/guide/how-it-works' },
              ],
            },
            {
              text: 'Usage',
              items: [
                { text: 'Features', link: '/en/guide/features' },
                { text: 'Configuration', link: '/en/guide/configuration' },
                { text: 'Scheduling', link: '/en/guide/schedule' },
                { text: 'CZeroX App', link: '/en/guide/app' },
              ],
            },
            {
              text: 'Advanced',
              items: [
                { text: 'Build from Source', link: '/en/guide/build' },
                { text: 'FAQ', link: '/en/guide/faq' },
              ],
            },
          ],
        },
        footer: {
          message: 'Released under the <a href="https://github.com/Xocio/CZero/blob/main/LICENSE" target="_blank" rel="noopener">GPL-3.0</a> License',
          copyright: 'Copyright © 2025 Xocio',
        },
      },
    },
  },
})
