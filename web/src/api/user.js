import request from './request'

export const register = (data) => request.post('/user/register', data)
export const login = (data) => request.post('/user/login', data)
export const getUserInfo = () => request.get('/user/info')
export const updateUser = (data) => request.put('/user/update', data)
export const searchUser = (keyword) => request.get('/user/search', { params: { keyword } })
