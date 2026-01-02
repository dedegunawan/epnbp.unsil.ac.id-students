import axios from 'axios'

function joinUrl(base: string, path: string): string {
    return `${base.replace(/\/+$/, '')}/${path.replace(/^\/+/, '')}`;
}

const baseURL = joinUrl(import.meta.env.VITE_BASE_URL || '', '/api')

export const api = axios.create({
    baseURL,
})

export default api



