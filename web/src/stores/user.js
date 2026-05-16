import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as userApi from '../api/user'
import { connect, disconnect } from '../ws'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const userInfo = ref(JSON.parse(localStorage.getItem('userInfo') || 'null'))

  const isLoggedIn = computed(() => !!token.value)

  async function loginReq(data) {
    const res = await userApi.login(data)
    token.value = res.data.token
    userInfo.value = res.data.user
    localStorage.setItem('token', res.data.token)
    localStorage.setItem('userInfo', JSON.stringify(res.data.user))
    connect(res.data.token)
    return res
  }

  async function registerReq(data) {
    return userApi.register(data)
  }

  async function fetchUserInfo() {
    const res = await userApi.getUserInfo()
    userInfo.value = res.data
    localStorage.setItem('userInfo', JSON.stringify(res.data))
  }

  async function updateProfile(data) {
    await userApi.updateUser(data)
    await fetchUserInfo()
  }

  function logout() {
    token.value = ''
    userInfo.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('userInfo')
    disconnect()
  }

  return { token, userInfo, isLoggedIn, loginReq, registerReq, fetchUserInfo, updateProfile, logout }
})
