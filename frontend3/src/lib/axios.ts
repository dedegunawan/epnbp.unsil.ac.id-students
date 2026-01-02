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

// Debug: log API configuration
if (import.meta.env.DEV) {
  console.log('API Configuration:', {
    VITE_API_URL: import.meta.env.VITE_API_URL,
    baseURL,
    VITE_BASE_URL: import.meta.env.VITE_BASE_URL,
  });
}

export const api = axios.create({
    baseURL,
})

// Add request interceptor to log requests
api.interceptors.request.use((config) => {
  const fullUrl = config.baseURL && config.url 
    ? (config.baseURL.endsWith('/') && config.url.startsWith('/') 
        ? config.baseURL + config.url.slice(1)
        : config.baseURL + config.url)
    : config.url;
  
  if (import.meta.env.DEV) {
    console.log('API Request:', config.method?.toUpperCase(), fullUrl);
  }
  return config;
});

export default api
