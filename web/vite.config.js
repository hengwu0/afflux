import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': {
        target: 'http://10.2.41.113:8080', //服务地址
        changeOrigin: true,
        rewrite: path => path.replace(/^\/api/, '')
      } 
    }
  }
})
