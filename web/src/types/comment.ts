export interface CommentAuthor {
  uid: string
  nickname: string
  avatarUrl: string
}

export interface Comment {
  uid: string
  author: CommentAuthor
  postUid: string
  rootUid: string
  parentUid?: string
  replyToAuthor?: CommentAuthor
  content: string
  images: string[]
  replyCount: number
  createdAt: string
  updatedAt: string
  likeCount: number
  liked: boolean
}
