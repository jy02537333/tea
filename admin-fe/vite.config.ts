import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0',
    port: 9094,
    proxy: {
      // 将前端 /api/* 请求代理到后端，目标地址取自 VITE_API_BASE（默认 127.0.0.1:9292）
      '/api': {
        target: process.env.VITE_API_BASE || 'http://127.0.0.1:9292',
        changeOrigin: true,
        // 如后端不需要重写，可直接返回原路径；此处保持直通
        rewrite: (path) => path,
      },
    },
  },
});
