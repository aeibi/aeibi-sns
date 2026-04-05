import {
  commentServiceGetComment,
  type CommentListRepliesResponse,
  useCommentServiceDeleteComment,
  useCommentServiceLikeComment,
  useCommentServiceListReplies,
} from "@/api/generated"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Card, CardContent } from "@/components/ui/card"
import { cn, formatCount, formatDateTime } from "@/lib/utils"
import type { Comment } from "@/types/comment"
import type { User } from "@/types/user"
import { keepPreviousData, useQueryClient } from "@tanstack/react-query"
import { ChevronLeftIcon, ChevronRightIcon, FlagIcon, MoreHorizontalIcon, ThumbsUpIcon, Trash2Icon } from "lucide-react"
import { useState } from "react"
import { PostReplyComposer } from "@/components/post-reply-composer"
import { PostReply } from "@/components/post-reply"
import { Button } from "@/components/ui/button"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { Spinner } from "@/components/ui/spinner"
import { PostCommentMedia } from "@/components/post-comment-media"
import { PostCommentText } from "@/components/post-comment-text"
import { toast } from "sonner"
import { Link } from "react-router-dom"

type PostCommentProps = React.ComponentProps<"div"> & {
  comment: Comment
  user?: User
  onUpdateComment: (patch: Partial<Comment>) => void
  onRemoveComment: () => void
}

export function PostComment({ comment, user, onUpdateComment, onRemoveComment }: PostCommentProps) {
  const isOwnComment = !!user && user.uid === comment.author.uid
  const queryClient = useQueryClient()
  const pageSize = 10
  const [showReply, setShowReply] = useState(false)
  const [page, setPage] = useState(1)
  const {
    data,
    queryKey,
    isFetching: isRepliesFetching,
    isPlaceholderData: isRepliesPlaceholderData,
  } = useCommentServiceListReplies(
    comment.uid,
    {
      page,
    },
    {
      query: { placeholderData: keepPreviousData, enabled: showReply },
    }
  )
  const replies = data?.comments ?? []
  const total = data?.total ?? 0
  const totalPage = Math.ceil(total / pageSize)

  const liked = comment.liked
  const likeCount = comment.likeCount
  const { mutate: likeComment } = useCommentServiceLikeComment()
  const { mutate: deleteComment, isPending: isDeletingComment } = useCommentServiceDeleteComment()
  const handleLike = () => {
    const previous = { liked, likeCount }
    const next = {
      liked: !liked,
      likeCount: Math.max(0, likeCount + (liked ? -1 : 1)),
    }
    onUpdateComment(next)
    likeComment(
      {
        uid: comment.uid,
        data: { uid: comment.uid, action: Number(next.liked) },
      },
      {
        onError: () => {
          toast.error("Failed to update like status", { position: "top-center" })
          onUpdateComment(previous)
        },
      }
    )
  }

  const handleShowReplies = () => {
    setPage(1)
    setShowReply((current) => !current)
  }

  const handleReplyPosted = (replyUid: string) => {
    setShowReply(true)
    onUpdateComment({ replyCount: comment.replyCount + 1 })
    void commentServiceGetComment(replyUid)
      .then((data) => {
        const nextReply = data.comment
        queryClient.setQueryData<CommentListRepliesResponse>(queryKey, (old) => {
          if (!old) {
            return {
              comments: [nextReply],
              page,
              total: Math.max(1, comment.replyCount + 1),
            }
          }
          if (old.comments.some((reply) => reply.uid === nextReply.uid)) return old
          return {
            ...old,
            comments: [...old.comments, nextReply],
            total: old.total + 1,
          }
        })
      })
      .catch(() => {
        toast.error("Failed to load created reply.", { position: "top-center" })
      })
  }

  const handleNextPage = () => {
    if (page < totalPage) setPage((current) => current + 1)
  }

  const handlePrevPage = () => {
    if (page > 1) setPage((current) => current - 1)
  }

  const handleRemoveReply = (replyUid: string) => {
    if (isDeletingComment) return
    const previousReplyCount = comment.replyCount
    onUpdateComment({ replyCount: Math.max(0, previousReplyCount - 1) })
    deleteComment(
      { uid: replyUid },
      {
        onSuccess: () => {
          queryClient.setQueryData<CommentListRepliesResponse>(queryKey, (old) => {
            if (!old) return old
            const comments = old.comments.filter((reply) => reply.uid !== replyUid)
            if (comments.length === old.comments.length) return old
            return {
              ...old,
              comments,
              total: Math.max(0, old.total - 1),
            }
          })
          if (page > 1 && replies.length === 1) {
            setPage((current) => Math.max(1, current - 1))
          }
        },
        onError: () => {
          onUpdateComment({ replyCount: previousReplyCount })
          toast.error("Failed to delete the reply.", { position: "top-center" })
        },
      }
    )
  }

  const handleRemoveComment = () => {
    if (isDeletingComment) return
    deleteComment(
      { uid: comment.uid },
      {
        onSuccess: onRemoveComment,
        onError: () => {
          toast.error("Failed to delete the comment.", { position: "top-center" })
        },
      }
    )
  }

  return (
    <div className="flex flex-col gap-2">
      <div className="group flex gap-2">
        <Link to={`/profile?uid=${encodeURIComponent(comment.author.uid)}`}>
          <Avatar>
            <AvatarImage src={comment.author.avatarUrl} alt={comment.author.nickname} />
            <AvatarFallback />
          </Avatar>
        </Link>
        <div className="flex min-w-0 flex-1 flex-col gap-1">
          <div className="flex items-start justify-between gap-2">
            <Link to={`/profile?uid=${encodeURIComponent(comment.author.uid)}`}>
              <p className="text-sm font-semibold hover:underline">{comment.author.nickname}</p>
            </Link>
            <span className="shrink-0 text-xs text-muted-foreground opacity-0 transition-opacity group-focus-within:opacity-100 group-hover:opacity-100">
              {formatDateTime(comment.createdAt)}
            </span>
          </div>
          <PostCommentText text={comment.content} />
          <PostCommentMedia images={comment.images} />
          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="xs"
              className={cn("text-muted-foreground hover:text-foreground", liked && "text-primary hover:text-primary")}
              onClick={handleLike}
            >
              <ThumbsUpIcon className="size-3" />
              <span>{formatCount(likeCount)}</span>
            </Button>
            <PostReplyComposer parentUid={comment.uid} isLoggedIn={!!user} onPosted={handleReplyPosted} />
            {(!!total || !!comment.replyCount) && (
              <div>
                <Button variant="ghost" size="xs" className="text-muted-foreground hover:text-foreground" onClick={handleShowReplies}>
                  <span>{showReply ? "hide reply" : "show reply"}</span>
                </Button>
              </div>
            )}
            <DropdownMenu>
              <DropdownMenuTrigger
                render={
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon-xs"
                    aria-label="More options"
                    className="ml-auto text-muted-foreground hover:text-foreground"
                  />
                }
              >
                <MoreHorizontalIcon className="size-3" />
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                {!isOwnComment && (
                  <DropdownMenuItem>
                    <FlagIcon className="text-muted-foreground" />
                    <span>Report</span>
                  </DropdownMenuItem>
                )}
                {isOwnComment && (
                  <DropdownMenuItem variant="destructive" disabled={isDeletingComment} onClick={handleRemoveComment}>
                    <Trash2Icon />
                    <span>Delete</span>
                  </DropdownMenuItem>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
      </div>

      {showReply && (
        <Card className="relative ml-8 flex flex-col gap-4 bg-muted/20 py-2">
          <CardContent className="px-2">
            <>
              {replies.map((reply) => (
                <PostReply key={reply.uid} comment={reply} user={user} onPosted={handleReplyPosted} onRemoveComment={handleRemoveReply} />
              ))}
              {page && totalPage > 1 && (
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon-xs"
                    className="text-muted-foreground hover:text-foreground"
                    onClick={handlePrevPage}
                  >
                    <ChevronLeftIcon />
                  </Button>
                  <span>
                    {page} / {totalPage}
                  </span>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon-xs"
                    className="text-muted-foreground hover:text-foreground"
                    onClick={handleNextPage}
                  >
                    <ChevronRightIcon />
                  </Button>
                </div>
              )}
            </>
          </CardContent>
          {isRepliesFetching && isRepliesPlaceholderData && !!replies.length && (
            <div className="absolute inset-0 flex items-center justify-center rounded-lg bg-background/70">
              <Spinner className="size-4 text-muted-foreground" />
            </div>
          )}
        </Card>
      )}
    </div>
  )
}
