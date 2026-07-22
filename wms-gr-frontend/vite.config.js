import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    proxy: {
      '/api/auth': { target: 'http://localhost:3002', changeOrigin: true },
      '/api/receiving': { target: 'http://localhost:3000', changeOrigin: true },
      '/api/shopee': { target: 'http://localhost:3000', changeOrigin: true },
      '/api/locations': { target: 'http://localhost:3003', changeOrigin: true },
      '/api/products': { target: 'http://localhost:3003', changeOrigin: true },
      '/api/stock': { target: 'http://localhost:3001', changeOrigin: true },
      '/api/report/inventory': { target: 'http://localhost:3001', changeOrigin: true },
      '/api/report/locations': { target: 'http://localhost:3003', changeOrigin: true },
      '/api/report/products': { target: 'http://localhost:3003', changeOrigin: true },
    },
  },
})
