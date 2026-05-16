import request from './request'

export const applyFriend = (data) => request.post('/friend/apply', data)
export const handleFriend = (data) => request.post('/friend/handle', data)
export const getApplications = (type) => request.get('/friend/applications', { params: { type } })
export const getFriendList = () => request.get('/friend/list')
export const deleteFriend = (friendId) => request.delete('/friend/delete', { params: { friend_id: friendId } })
