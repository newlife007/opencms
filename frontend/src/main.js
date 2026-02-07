import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import enUS from 'element-plus/es/locale/lang/en'

import App from './App.vue'
import router from './router'
import i18n, { getLocale } from './i18n'

const app = createApp(App)

// Create Pinia instance first (before router to avoid circular dependency)
const pinia = createPinia()
app.use(pinia)

// Use i18n
app.use(i18n)

// Then use router
app.use(router)

// Use Element Plus with dynamic locale
const currentLocale = getLocale()
const elementLocale = currentLocale === 'en-US' ? enUS : zhCn
app.use(ElementPlus, {
  locale: elementLocale,
})

// Register Element Plus icons asynchronously to avoid blocking
import('@element-plus/icons-vue').then((icons) => {
  for (const [key, component] of Object.entries(icons)) {
    app.component(key, component)
  }
})

app.mount('#app')
