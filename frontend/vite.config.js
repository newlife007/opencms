import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    host: '0.0.0.0', // 监听所有网络接口，允许外网访问
    port: 3000,
    // 开发环境禁用缓存
    headers: {
      'Cache-Control': 'no-store, no-cache, must-revalidate, proxy-revalidate',
      'Pragma': 'no-cache',
      'Expires': '0',
    },
    // API代理配置 - 开发环境必需
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    sourcemap: false,
    rollupOptions: {
      output: {
        manualChunks(id) {
          // Separate chunks to avoid circular dependencies
          if (id.includes('node_modules')) {
            // Element Plus and icons in separate chunks
            if (id.includes('element-plus')) {
              return 'element-plus'
            }
            if (id.includes('@element-plus/icons-vue')) {
              return 'element-icons'
            }
            // Vue core
            if (id.includes('vue/')) {
              return 'vue-core'
            }
            // Vue Router separate from Pinia to avoid circular dependency
            if (id.includes('vue-router')) {
              return 'vue-router'
            }
            // Pinia separate
            if (id.includes('pinia')) {
              return 'pinia'
            }
            // Video.js core - separate from plugins to avoid circular deps
            if (id.includes('video.js') && !id.includes('videojs-')) {
              return 'videojs-core'
            }
            // Video.js plugins and FLV.js - separate chunk
            if (id.includes('videojs-') || id.includes('flv.js')) {
              return 'videojs-plugins'
            }
            // Other vendor libs
            return 'vendor'
          }
        },
      },
    },
  },
})
