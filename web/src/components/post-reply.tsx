import { cn, formatCount } from "@/lib/utils"
import type { Comment } from "@/types/comment"
import type { User } from "@/types/user"
import { FlagIcon, MoreHorizontalIcon, ThumbsUpIcon, Trash2Icon } from "lucide-react"
import { PostReplyComposer } from "@/components/post-reply-composer"
import { Button } from "@/components/ui/button"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { useState } from "react"
import { useCommentServiceLikeComment } from "@/api/generated"
import { toast } from "sonner"
import { Link } from "react-router-dom"

type PostReplyProps = React.ComponentProps<"div"> & {
  comment: Comment
  user?: User
  onPosted?: (commentUid: string) => void
  onRemoveComment?: (commentUid: string) => void
}

export function PostReply({ comment, user, onPosted, onRemoveComment }: PostReplyProps) {
  const isOwnComment = !!user && user.uid === comment.author.uid

  const [likeStatus, setLikeStatus] = useState({ liked: comment.liked, likeCount: comment.likeCount })
  const { mutate: likeComment } = useCommentServiceLikeComment()
  const handleLike = () => {
    setLikeStatus((current) => {
      return { liked: !current.liked, likeCount: current.likeCount + (likeStatus.liked ? -1 : 1) }
    })
    likeComment(
      {
        uid: comment.uid,
        data: { uid: comment.uid, action: Number(!likeStatus.liked) },
      },
      {
        onError: () => {
          toast.error("Failed to update like status", { position: "top-center" })
          setLikeStatus((current) => {
            return { liked: !current.liked, likeCount: current.likeCount + (likeStatus.liked ? -1 : 1) }
          })
        },
      }
    )
  }

  return (
    <div className="flex gap-2">
      <div className="flex min-w-0 flex-1 flex-col gap-1">
        <p className="text-sm leading-relaxed wrap-break-word text-foreground/90">
          <Link to={`/profile?uid=${encodeURIComponent(comment.author.uid)}`}>
            <span className="font-semibold text-foreground hover:underline">{comment.author.nickname}</span>
          </Link>
          {comment.replyToAuthor && (
            <>
              <span> reply </span>
              <Link to={`/profile?uid=${encodeURIComponent(comment.replyToAuthor.uid)}`}>
                <span className="font-semibold text-foreground hover:underline">{comment.replyToAuthor.nickname}</span>
              </Link>
            </>
          )}
          <span> : {comment.content}</span>
        </p>
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="xs"
            className={cn("text-muted-foreground hover:text-foreground", likeStatus.liked && "text-primary hover:text-primary")}
            onClick={handleLike}
          >
            <ThumbsUpIcon className="size-3" />
            <span>{formatCount(likeStatus.likeCount)}</span>
          </Button>
          <PostReplyComposer parentUid={comment.uid} isLoggedIn={!!user} onPosted={(uid) => onPosted?.(uid)} />
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
                <DropdownMenuItem variant="destructive" onClick={() => onRemoveComment?.(comment.uid)}>
                  <Trash2Icon />
                  <span>Delete</span>
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </div>
  )
}
