import DefaultTheme from 'vitepress/theme'
import type { Theme } from 'vitepress'
import DownloadEditions from './components/DownloadEditions.vue'
import './custom.css'

export default {
  extends: DefaultTheme,
  enhanceApp({ app }) {
    app.component('DownloadEditions', DownloadEditions)
  },
} satisfies Theme
