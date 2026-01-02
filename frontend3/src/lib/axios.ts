import axios from 'axios'

// Use VITE_API_URL from environment
// Production: https://epnbp.unsil.ac.id/students/api
// Docker/Local: /api (nginx will proxy to backend)
// If VITE_API_URL is absolute URL (starts with http), use it directly
// Otherwise, use relative path (nginx will handle it)
const apiUrl = import.meta.env.VITE_API_URL || '/api';
const baseURL = apiUrl.startsWith('http') 
  ? apiUrl 
  : apiUrl; // Use as-is for relative paths (nginx will handle it)

export const api = axios.create({
    baseURL,
})

export default api
