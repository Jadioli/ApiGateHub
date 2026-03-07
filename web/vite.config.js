import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      '/admin': 'http://localhost:9011',
      '/v1': 'http://localhost:9011',
      '/anthropic': 'http://localhost:9011',
    },
  },
  build: {
    outDir: 'dist',
  },
})

