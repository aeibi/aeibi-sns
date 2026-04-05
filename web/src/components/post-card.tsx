import type { Post } from "@/types/post"
import type { User } from "@/types/user"
import {
  commentServiceGetComment,
  type CommentListTopCommentsResponse,
  getCommentServiceListTopCommentsQueryKey,
  getUserServiceGetUserQueryKey,
  useFollowServiceFollow,
  useCommentServiceListTopComments,
  usePostServiceCollectPost,
  usePostServiceDeletePost,
  usePostServiceLikePost,
} from "@/api/generated"
import { useQueryClient } from "@tanstack/react-query"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { BookmarkIcon, HeartIcon, MessageCircleIcon } from "lucide-react"
import { useState } from "react"
import { Link, useLocation } from "react-router-dom"
import { PostActionButton } from "@/components/post-action"
import { PostComment } from "@/components/post-comment"
import { PostCommentsComposer } from "@/components/post-comment-composer"
import { PostHeader } from "@/components/post-header"
import { PostMedia } from "@/components/post-media"
import { PostTag } from "@/components/post-tag"
import { PostText } from "@/components/post-text"
import type { Comment } from "@/types/comment"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"

type PostCardProps = {
  post: Post
  user?: User
  onUpdatePost: (patch: Partial<Post>) => void
  onRemovePost: () => void
  disableCommentExpand?: boolean
}

export function PostCard({ post, user, onUpdatePost, onRemovePost, disableCommentExpand = false }: PostCardProps) {
  const isOwnPost = !!user && user.uid === post.author.uid
  const [commentActive, setCommentActive] = useState(false)
  const canExpandComments = !disableCommentExpand
  const queryClient = useQueryClient()
  const { search } = useLocation()

  const { mutate: deletePost, isPending: isDeleting } = usePostServiceDeletePost()
  const { mutate: likePost, isPending: isLiking } = usePostServiceLikePost()
  const { mutate: collectPost, isPending: isCollecting } = usePostServiceCollectPost()
  const { mutate: followUser, isPending: isFollowPending } = useFollowServiceFollow()

  const previewCommentsQueryKey = [...getCommentServiceListTopCommentsQueryKey(post.uid), "preview"] as const
  const { data, refetch, queryKey } = useCommentServiceListTopComments(post.uid, undefined, {
    query: {
      enabled: canExpandComments && commentActive,
      queryKey: previewCommentsQueryKey,
    },
  })
  const comments = data?.comments ?? []

  const handleUpdateComment = (uid: string, patch: Partial<Comment>) => {
    queryClient.setQueryData(queryKey, (old) => {
      if (!old) return old
      let changed = false
      const comments = old.comments.map((comment) => {
        if (comment.uid !== uid) return comment
        changed = true
        return { ...comment, ...patch }
      })
      return changed ? { ...old, comments } : old
    })
  }

  const handleRemoveComment = () => {
    refetch()
  }

  const handleCommentPosted = (uid: string) => {
    onUpdatePost({ commentCount: post.commentCount + 1 })
    void commentServiceGetComment(uid)
      .then((data) => {
        queryClient.setQueryData<CommentListTopCommentsResponse>(queryKey, (old) => {
          if (!old) return old
          if (old.comments.some((comment) => comment.uid === data.comment.uid)) return old
          return { ...old, comments: [data.comment, ...old.comments] }
        })
      })
      .catch(() => {
        toast.error("Failed to load created comment.", { position: "top-center" })
      })
  }

  const handleCommentClick = () => {
    if (!canExpandComments) return
    setCommentActive((active) => !active)
  }

  const handleDelete = () => {
    if (isDeleting) return
    deletePost(
      { uid: post.uid },
      {
        onSuccess: onRemovePost,
        onError: () => toast.error("Failed to delete the post.", { position: "top-center" }),
      }
    )
  }

  const handleFollow = () => {
    if (!user || isOwnPost || isFollowPending) return
    const previousAuthor = post.author
    onUpdatePost({
      author: {
        ...previousAuthor,
        isFollowing: !previousAuthor.isFollowing,
      },
    })
    followUser(
      {
        uid: previousAuthor.uid,
        data: { uid: previousAuthor.uid, action: Number(previousAuthor.isFollowing) },
      },
      {
        onSuccess: () => {
          void queryClient.invalidateQueries({ queryKey: getUserServiceGetUserQueryKey(previousAuthor.uid) })
        },
        onError: () => {
          toast.error("Failed to update follow status.", { position: "top-center" })
          onUpdatePost({ author: previousAuthor })
        },
      }
    )
  }

  const handleLike = () => {
    if (!user || isLiking) return
    onUpdatePost({
      liked: !post.liked,
      likeCount: Math.max(0, post.likeCount + (post.liked ? -1 : 1)),
    })
    likePost(
      {
        uid: post.uid,
        data: { uid: post.uid, action: Number(post.liked) },
      },
      {
        onError: () => {
          toast.error("Failed to like the post.", { position: "top-center" })
          onUpdatePost({
            liked: post.liked,
            likeCount: post.likeCount,
          })
        },
      }
    )
  }

  const handleCollect = () => {
    if (!user || isCollecting) return
    onUpdatePost({
      collected: !post.collected,
      collectionCount: Math.max(0, post.collectionCount + (post.collected ? -1 : 1)),
    })
    collectPost(
      {
        uid: post.uid,
        data: { uid: post.uid, action: Number(post.collected) },
      },
      {
        onError: () => {
          toast.error("Failed to collect the post.", { position: "top-center" })
          onUpdatePost({
            collected: post.collected,
            collectionCount: post.collectionCount,
          })
        },
      }
    )
  }

  return (
    <Card>
      <CardHeader>
        <PostHeader post={post} isOwnPost={isOwnPost} onDelete={handleDelete} onFollow={handleFollow} />
      </CardHeader>
      <CardContent className="flex flex-col gap-3">
        <PostText post={post} />
        <PostMedia post={post} />
        <PostTag tags={post.tags} />
        <div className="flex items-center gap-1 text-muted-foreground">
          <PostActionButton
            icon={<MessageCircleIcon />}
            count={post.commentCount}
            active={canExpandComments && commentActive}
            disabled={!canExpandComments}
            onClick={() => handleCommentClick()}
          />
          <PostActionButton
            icon={<HeartIcon />}
            count={post.likeCount}
            active={post.liked}
            disabled={!user || isLiking}
            onClick={handleLike}
          />
          <PostActionButton
            icon={<BookmarkIcon />}
            count={post.collectionCount}
            active={post.collected}
            disabled={!user || isCollecting}
            onClick={handleCollect}
          />
        </div>
        {canExpandComments && commentActive && (
          <div className="flex flex-col gap-4">
            {!!user && <PostCommentsComposer user={user} postUid={post.uid} onPosted={handleCommentPosted} />}
            {comments.length > 0 && (
              <>
                {comments.slice(0, 5).map((comment) => (
                  <PostComment
                    key={comment.uid}
                    comment={comment}
                    user={user}
                    onUpdateComment={(patch) => handleUpdateComment(comment.uid, patch)}
                    onRemoveComment={handleRemoveComment}
                  />
                ))}
                {comments.length > 5 && (
                  <Button
                    variant="link"
                    size="sm"
                    className="m-auto h-auto w-fit p-0 text-muted-foreground hover:text-foreground"
                    render={<Link to={{ pathname: `post/${post.uid}`, search }} />}
                  >
                    Show More
                  </Button>
                )}
              </>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
