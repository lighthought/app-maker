import { createI18n } from 'vue-i18n'
import zh from './zh'
import en from './en'

export type MessageSchema = typeof zh

const messages = {
  zh,
  en
}

const i18n = createI18n({
  legacy: false, // 使用 Composition API 模式
  locale: 'zh', // 默认语言
  fallbackLocale: 'en', // 回退语言
  messages,
  globalInjection: true, // 全局注入 $t 方法
  silentTranslationWarn: true, // 禁用警告
  silentFallbackWarn: true // 禁用回退警告
})

export default i18n

// 导出常用的方法
export const t = i18n.global.t
export const locale = i18n.global.locale
export const setLocale = (lang: 'zh' | 'en') => {
  i18n.global.locale.value = lang
  localStorage.setItem('preferred-language', lang)
}

// 初始化语言设置
export const initLanguage = () => {
  const savedLanguage = localStorage.getItem('preferred-language') as 'zh' | 'en'
  if (savedLanguage && ['zh', 'en'].includes(savedLanguage)) {
    i18n.global.locale.value = savedLanguage
  } else {
    // 根据浏览器语言自动检测
    const browserLanguage = navigator.language.toLowerCase()
    if (browserLanguage.startsWith('zh')) {
      i18n.global.locale.value = 'zh'
    } else {
      i18n.global.locale.value = 'en'
    }
    localStorage.setItem('preferred-language', i18n.global.locale.value)
  }
}

// 切换语言方法
export const toggleLanguage = () => {
  const currentLang = i18n.global.locale
  const newLang = currentLang.value === 'zh' ? 'en' : 'zh'
  setLocale(newLang)
  return newLang
}

// 获取当前语言
export const getCurrentLanguage = () => {
  return i18n.global.locale
}