import axios from 'axios'

// Gunakan import.meta.env agar bisa akses variabel environment dari Vite
const baseURL = `${import.meta.env.VITE_BASE_URL}/api`

export const api = axios.create({
    baseURL,
})

export default api
