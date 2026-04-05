export type User = {
  uid: string
  username?: string
  role: string
  email?: string
  nickname: string
  avatarUrl: string
  followersCount: number
  followingCount: number
  isFollowing?: boolean
  description?: string
}
