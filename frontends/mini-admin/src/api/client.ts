import axios from 'axios'

const API_URL = (import.meta as ImportMeta).env.VITE_API_URL || 'http://localhost/v1'
const INTERNAL_SERVICE_KEY = (import.meta as ImportMeta).env.VITE_INTERNAL_SERVICE_KEY || 'IlikeCookies'

export const apiClient = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
    ...(INTERNAL_SERVICE_KEY
      ? {
          'X-Internal-API-Key': INTERNAL_SERVICE_KEY,
        }
      : {}),
  },
})

// Request interceptor for auth token if needed
apiClient.interceptors.request.use(
  (config) => {
    // Add auth token here if required
    // const token = localStorage.getItem('token')
    // if (token) {
    //   config.headers.Authorization = `Bearer ${token}`
    // }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor for error handling
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    console.error('API Error:', error.response?.data || error.message)
    return Promise.reject(error)
  }
)
