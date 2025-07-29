import axios from 'axios'

// Gunakan import.meta.env agar bisa akses variabel environment dari Vite
const baseURL = joinUrl(import.meta.env.VITE_BASE_URL, '/api')

function joinUrl(base: string, path: string): string {
    return `${base.replace(/\/+$/, '')}/${path.replace(/^\/+/, '')}`;
}


export const api = axios.create({
    baseURL,
})

export default api
