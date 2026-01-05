import axios from 'axios'

// Use VITE_API_URL from environment
// Production: https://epnbp.unsil.ac.id/students/api
// Docker/Local: /api (nginx will proxy to backend)
// Development: gunakan VITE_API_URL dari .env file
const apiUrl = import.meta.env.VITE_API_URL;
const isDev = import.meta.env.DEV || import.meta.env.MODE === 'development';

let baseURL: string;
if (apiUrl && apiUrl.trim() !== '') {
  // Jika VITE_API_URL diset dan tidak kosong
  if (apiUrl.startsWith('http')) {
    // Absolute URL: gunakan langsung (contoh: http://localhost:8080/api)
    baseURL = apiUrl.endsWith('/') ? apiUrl.slice(0, -1) : apiUrl;
  } else {
    // Relative path: gunakan langsung (contoh: /api)
    baseURL = apiUrl.startsWith('/') ? apiUrl : '/' + apiUrl;
  }
} else {
  // Default jika VITE_API_URL tidak diset
  if (isDev) {
    // Development: default ke localhost:8080/api
    baseURL = 'http://localhost:8080/api';
  } else {
    // Production/Docker: gunakan relative path /api (nginx akan proxy)
    baseURL = '/api';
  }
}

// Pastikan baseURL tidak double slash di akhir
if (baseURL.endsWith('/') && baseURL.length > 1) {
  baseURL = baseURL.slice(0, -1);
}

// Debug: log API configuration
if (import.meta.env.DEV) {
  console.log('API Configuration:', {
    VITE_API_URL: import.meta.env.VITE_API_URL,
    baseURL,
    VITE_BASE_URL: import.meta.env.VITE_BASE_URL,
    mode: import.meta.env.MODE,
  });
}

export const api = axios.create({
    baseURL: baseURL,
    timeout: 30000, // 30 seconds timeout
})

// Debug: verify axios instance configuration
if (import.meta.env.DEV) {
  console.log('Axios Instance Created:', {
    baseURL: api.defaults.baseURL,
    timeout: api.defaults.timeout,
  });
}

// Add request interceptor for logging (no token needed for public endpoints)
api.interceptors.request.use((config) => {
  // Log request untuk debugging (tanpa token karena endpoint public)
  if (import.meta.env.DEV) {
    if (config.baseURL && config.url) {
      const base = config.baseURL.endsWith('/') ? config.baseURL.slice(0, -1) : config.baseURL;
      const path = config.url.startsWith('/') ? config.url : '/' + config.url;
      const fullUrl = base + path;
      console.log('API Request:', config.method?.toUpperCase(), fullUrl);
    } else {
      console.log('API Request:', config.method?.toUpperCase(), config.url);
    }
  }
  return config;
});

export default api
