import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { defineConfig } from 'vite'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  build: { outDir: '.tmp-dist', emptyOutDir: true },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      'vue-demi': fileURLToPath(new URL('./src/dev-shims/vue-demi.mjs', import.meta.url)),
    },
  },
})
