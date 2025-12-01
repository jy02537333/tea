import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

// Manual chunking example: put large libs into separate chunks
export default defineConfig({
  resolve: {
    alias: {
      '@local/ui-thumbnail/react': path.resolve(__dirname, '../packages/ui-thumbnail/dist/react/index.js')
    }
  },
  plugins: [react()],
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          // put antd into its own chunk
          antd: ['antd'],
          // react + react-dom in a vendors chunk
          vendor: ['react', 'react-dom', 'react-router-dom']
        }
      }
    },
    chunkSizeWarningLimit: 1200 // raise warning threshold to 1.2MB
  }
})
