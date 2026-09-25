import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  server: {
    host: '0.0.0.0',
    port: 53083,
    strictPort: true,
    proxy: {
      '/api': {
        target: process.env.FASTIMG_BACKEND_URL ?? 'http://172.29.160.1:53084',
        changeOrigin: true,
      },
    },
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      'vue-demi': fileURLToPath(new URL('./src/dev-shims/vue-demi.mjs', import.meta.url)),
    },
  },
})
