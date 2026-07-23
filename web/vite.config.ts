import { sveltekit } from '@sveltejs/kit/vite'
import tailwindcss from '@tailwindcss/vite'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [sveltekit(), tailwindcss()],
  build: {
    sourcemap: true,
  },
  define: {
    __VERSION__: JSON.stringify(process.env.npm_package_version),
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
      },
      '/nodeinfo': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
      },
    },
    watch: {
      ignored: ['!**/node_modules/mono-svelte/**'],
    },
  },
})
