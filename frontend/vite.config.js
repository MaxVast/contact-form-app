import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// Le port et le proxy /api sont utiles en dev local (npm run dev).
// En prod, c'est Nginx (voir nginx.conf) qui route /api vers le backend.
export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  },
  test: {
    environment: 'jsdom',
    globals: true
  }
})
