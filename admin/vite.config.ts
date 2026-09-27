import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath, URL } from 'node:url'

const backendTarget = process.env.FASTIMG_BACKEND_URL ?? 'http://127.0.0.1:53085'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(), tailwindcss()],
  server: {
    host: '0.0.0.0',
    port: 5180,
    strictPort: true,
    proxy: {
      '/api': {
        target: backendTarget,
        changeOrigin: true,
      },
      '/sitemap.xml': { target: backendTarget, changeOrigin: true },
      '/robots.txt': { target: backendTarget, changeOrigin: true },
    },
  },
  preview: {
    host: '0.0.0.0',
    port: 4173,
    strictPort: true,
    proxy: {
      '/api': { target: backendTarget, changeOrigin: true },
      '/sitemap.xml': { target: backendTarget, changeOrigin: true },
      '/robots.txt': { target: backendTarget, changeOrigin: true },
    },
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      'vue-demi': fileURLToPath(new URL('./src/dev-shims/vue-demi.mjs', import.meta.url)),
    },
  },
})
