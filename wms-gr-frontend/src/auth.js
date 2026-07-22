import { ref } from 'vue'
import axios from 'axios'

export const token = ref('')

export async function login(email, password) {
  const res = await axios.post('/api/auth/login', { email, password })
  token.value = res.data.token
}

export function logout() {
  token.value = ''
}

export function authHeader() {
  return { Authorization: `Bearer ${token.value}` }
}
