import { createI18n } from 'vue-i18n'
import zhCN from './locales/zh-CN.json'
import enUS from './locales/en-US.json'

// 获取浏览器语言或本地存储的语言设置
const getDefaultLocale = () => {
  const savedLocale = localStorage.getItem('locale')
  if (savedLocale) {
    return savedLocale
  }
  
  // 从浏览器获取语言设置
  const browserLocale = navigator.language || navigator.userLanguage
  if (browserLocale.startsWith('zh')) {
    return 'zh-CN'
  }
  return 'en-US'
}

const i18n = createI18n({
  legacy: false, // 使用 Composition API 模式
  locale: getDefaultLocale(),
  fallbackLocale: 'zh-CN',
  messages: {
    'zh-CN': zhCN,
    'en-US': enUS
  },
  globalInjection: true, // 全局注入 $t 函数
})

export default i18n

// 导出切换语言的函数
export const setLocale = (locale) => {
  i18n.global.locale.value = locale
  localStorage.setItem('locale', locale)
  // 更新 HTML lang 属性
  document.querySelector('html').setAttribute('lang', locale)
}

// 导出获取当前语言的函数
export const getLocale = () => {
  return i18n.global.locale.value
}

// 导出可用语言列表
export const availableLocales = [
  { value: 'zh-CN', label: '简体中文' },
  { value: 'en-US', label: 'English' }
]
