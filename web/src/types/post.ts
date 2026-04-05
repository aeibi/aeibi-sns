export interface PostAuthor {
  uid: string
  nickname: string
  avatarUrl: string
  isFollowing: boolean
}

export interface Attachment {
  url: string
  name: string
  size: string
  contentType: string
  checksum: string
}

export interface Post {
  uid: string
  author: PostAuthor
  text: string
  images: string[]
  attachments: Attachment[]
  tags: string[]
  commentCount: number
  collectionCount: number
  likeCount: number
  visibility: string
  latestRepliedOn: string
  ip: string
  pinned: boolean
  liked: boolean
  collected: boolean
  createdAt: string
  updatedAt: string
}
