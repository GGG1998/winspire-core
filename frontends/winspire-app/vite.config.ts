import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  preview: {
    allowedHosts: ['winspire-dev-s63lr.ondigitalocean.app, localhost', 'dev-api.gowinspire.com'],
  },
})
