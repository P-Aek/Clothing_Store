const API_BASE = import.meta.env.VITE_API_BASE || 'https://clothing-store-icuz.onrender.com/api'

export async function apiRequest(path, options = {}) {
  const token = localStorage.getItem('token')
  const headers = new Headers(options.headers || {})
  if (!headers.has('Content-Type') && options.body !== undefined) headers.set('Content-Type', 'application/json')
  if (token) headers.set('Authorization', `Bearer ${token}`)
  const response = await fetch(`${API_BASE}${path}`, { ...options, headers })
  const text = await response.text()
  let data = null
  if (text) {
    try { data = JSON.parse(text) } catch { data = { message: text } }
  }
  if (!response.ok) {
    const error = new Error(data?.message || data?.error || `Request failed (${response.status})`)
    error.status = response.status
    error.data = data
    throw error
  }
  return data
}

export function getErrorMessage(error, fallback = 'เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง') {
  return error?.data?.message || error?.data?.error || error?.message || fallback
}

export { API_BASE }
