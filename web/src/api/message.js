import request from './request'

export const getHistory = (params) => request.get('/message/history', { params })
export const getConversations = () => request.get('/message/conversations')
export const searchMessage = (params) => request.get('/message/search', { params })
export const uploadImage = (file) => {
  const form = new FormData()
  form.append('file', file)
  return request.post('/message/upload', form)
}
