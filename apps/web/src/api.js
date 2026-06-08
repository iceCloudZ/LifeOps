import axios from 'axios'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080',
  headers: {
    'Authorization': `Bearer ${import.meta.env.VITE_API_TOKEN || 'dev-token'}`
  }
})

export default api
