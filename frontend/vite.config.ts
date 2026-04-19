import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 3001,
    proxy: {
      // REST API 代理到后端 Gin HTTP（开发用）
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  // gRPC-Web 生成的代码用 CommonJS + google-protobuf 需要特殊处理
  optimizeDeps: {
    include: ['google-protobuf', '@improbable-eng/grpc-web'],
  },
  build: {
    commonjsOptions: {
      include: [/grpc-web/, /google-protobuf/, /node_modules/],
    },
  },
})
