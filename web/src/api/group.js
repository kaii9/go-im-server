import request from './request'

export const createGroup = (data) => request.post('/group/create', data)
export const joinGroup = (data) => request.post('/group/join', data)
export const leaveGroup = (data) => request.post('/group/leave', data)
export const getGroupInfo = (id) => request.get('/group/info', { params: { id } })
export const getGroupMembers = (id) => request.get('/group/members', { params: { id } })
export const getMyGroups = () => request.get('/group/mine')
export const inviteMember = (data) => request.post('/group/invite', data)
