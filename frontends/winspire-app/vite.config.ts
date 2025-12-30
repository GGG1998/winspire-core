import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    allowedHosts: [
      'localhost',
      'winspire-dev-s63lr.ondigitalocean.app',
      'dev-api.gowinspire.com',
    ],
  },
  preview: {
    allowedHosts: [
      'localhost',
      'winspire-dev-s63lr.ondigitalocean.app',
      'dev-api.gowinspire.com',
    ],
  },
})
