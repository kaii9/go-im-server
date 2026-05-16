import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as friendApi from '../api/friend'
import * as groupApi from '../api/group'

export const useFriendStore = defineStore('friend', () => {
  const friendList = ref([])
  const receivedApps = ref([])
  const sentApps = ref([])
  const myGroups = ref([])

  async function fetchFriends() {
    const res = await friendApi.getFriendList()
    friendList.value = res.data || []
  }

  async function fetchApplications() {
    const [recv, sent] = await Promise.all([
      friendApi.getApplications('received'),
      friendApi.getApplications('sent'),
    ])
    receivedApps.value = recv.data || []
    sentApps.value = sent.data || []
  }

  async function applyFriend(toUserId, reason) {
    await friendApi.applyFriend({ to_user_id: toUserId, reason })
    await fetchApplications()
  }

  async function handleApplication(applicationId, agree) {
    await friendApi.handleFriend({ application_id: applicationId, agree })
    await fetchApplications()
    await fetchFriends()
  }

  async function deleteFriend(friendId) {
    await friendApi.deleteFriend(friendId)
    await fetchFriends()
  }

  async function fetchMyGroups() {
    const res = await groupApi.getMyGroups()
    myGroups.value = res.data || []
  }

  async function createGroup(name) {
    await groupApi.createGroup({ name })
    await fetchMyGroups()
  }

  async function joinGroup(groupId) {
    await groupApi.joinGroup({ group_id: groupId })
    await fetchMyGroups()
  }

  async function leaveGroup(groupId) {
    await groupApi.leaveGroup({ group_id: groupId })
    await fetchMyGroups()
  }

  async function inviteGroupMember(groupId, userId) {
    await groupApi.inviteMember({ group_id: groupId, user_id: userId })
    await fetchMyGroups()
  }

  return {
    friendList, receivedApps, sentApps, myGroups,
    fetchFriends, fetchApplications, applyFriend, handleApplication, deleteFriend,
    fetchMyGroups, createGroup, joinGroup, leaveGroup, inviteGroupMember,
  }
})
