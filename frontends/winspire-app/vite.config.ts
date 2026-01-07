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
    proxy: {
      // Game Management service
      '/v1/games': {
        target: 'http://localhost:8088',
        changeOrigin: true,
      },
      // Tournament service (matches /v1/{uuid}/tournaments)
      '^/v1/[0-9a-f-]{36}/tournaments': {
        target: 'http://localhost:8088',
        changeOrigin: true,
      },
      // Matchmaking service
      '/v1/matchmaking': {
        target: 'http://localhost:8088',
        changeOrigin: true,
        ws: true,
      },
    },
  },
  preview: {
    allowedHosts: [
      'localhost',
      'winspire-dev-s63lr.ondigitalocean.app',
      'dev-api.gowinspire.com',
    ],
  },
})
