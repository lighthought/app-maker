import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import i18n, { initLanguage } from '@/locales'
import './styles/main.scss'
import './userWorker'

const app = createApp(App)

// 初始化语言设置
initLanguage()

app.use(createPinia())
app.use(router)
app.use(i18n)

app.mount('#app')