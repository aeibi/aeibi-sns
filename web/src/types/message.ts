export interface MessageActor {
  uid: string
  nickname: string
  avatarUrl: string
}

export interface CommentMessage {
  uid: string
  isRead: boolean
  actor: MessageActor
  createdAt: string
  commentUid?: string
  commentContent?: string
  postUid?: string
  parentUid?: string
  parentContent?: string
}

export interface FollowMessage {
  uid: string
  isRead: boolean
  actor: MessageActor
  createdAt: string
}

export type InboxMessage = FollowMessage | CommentMessage
