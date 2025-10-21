import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [
    vue()
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  server: {
    port: 3000,
    host: true,
    proxy: {
      // API 代理到本地后端服务
      '/api': {
        target: 'http://localhost:8098',
        changeOrigin: true,
        secure: false,
      },
      // WebSocket 代理到本地后端服务
      '/ws': {
        target: 'http://localhost:8098',
        changeOrigin: true,
        ws: true,
      },
    },
  },
  // 可选：强制预构建某些 Worker 以避免开发阶段的问题
  optimizeDeps: {
    include: [
      `monaco-editor/esm/vs/language/json/json.worker`,
      `monaco-editor/esm/vs/language/css/css.worker`,
      `monaco-editor/esm/vs/language/html/html.worker`,
      `monaco-editor/esm/vs/language/typescript/ts.worker`,
      `monaco-editor/esm/vs/editor/editor.worker`
    ]
  },
  build: {
    target: 'es2015',
    outDir: 'dist',
    assetsDir: 'assets',
    sourcemap: true,
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['vue', 'vue-router', 'pinia'],
          ui: ['naive-ui'],
          utils: ['axios', '@vueuse/core']
        }
      }
    }
  }
})