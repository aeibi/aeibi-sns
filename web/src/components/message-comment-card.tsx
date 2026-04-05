import { useState } from "react"
import { ThumbsUpIcon } from "lucide-react"
import { useCommentServiceLikeComment } from "@/api/generated"
import { Avatar, AvatarBadge, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Card, CardContent } from "@/components/ui/card"
import type { CommentMessage } from "@/types/message"
import { Button } from "@/components/ui/button"
import { Link } from "react-router-dom"
import { cn, formatDateTime } from "@/lib/utils"
import { toast } from "sonner"
import { PostReplyComposer } from "./post-reply-composer"

type MessageCommentCardProps = {
  message: CommentMessage
}

export function MessageCommentCard({ message }: MessageCommentCardProps) {
  const [liked, setLiked] = useState(false)
  const { mutate: likeComment, isPending: isLiking } = useCommentServiceLikeComment()
  const action = message.parentUid === message.postUid ? "commented on your post." : "replied to your comment."
  if (!message.postUid) return null
  if (!message.commentUid) return null

  const handleLike = () => {
    if (isLiking) return
    if (!message.commentUid) return
    const previousLiked = liked
    const nextLiked = !previousLiked
    setLiked(nextLiked)
    likeComment(
      {
        uid: message.commentUid,
        data: { uid: message.commentUid, action: Number(nextLiked) },
      },
      {
        onError: () => {
          toast.error("Failed to update like status", { position: "top-center" })
          setLiked(previousLiked)
        },
      }
    )
  }

  return (
    <Card>
      <CardContent className="flex gap-3">
        <Link to="/">
          <Avatar size="lg">
            <AvatarImage src={message.actor.avatarUrl} alt={message.actor.nickname} />
            <AvatarFallback />
            {!message.isRead && <AvatarBadge className="bg-red-300" />}
          </Avatar>
        </Link>
        <div className="flex flex-1 flex-col gap-3">
          <p className="flex gap-2 text-sm">
            <Link to="/" className="font-semibold hover:underline">
              {message.actor.nickname}
            </Link>
            <span>{action}</span>
          </p>

          {!!message.commentContent && <p className="text-sm wrap-break-word">{message.commentContent}</p>}
          {!!message.parentContent && <p className="text-xs wrap-break-word text-muted-foreground">Replying to: {message.parentContent}</p>}
          <div className="flex items-center gap-3 text-muted-foreground">
            <span className="text-xs text-muted-foreground">{formatDateTime(message.createdAt)}</span>
            <Button
              variant="ghost"
              size="icon-sm"
              aria-pressed={liked}
              disabled={isLiking}
              onClick={handleLike}
              className={cn("text-muted-foreground hover:text-foreground", liked && "text-primary hover:text-primary")}
            >
              <ThumbsUpIcon />
            </Button>
            <PostReplyComposer parentUid={message.commentUid} isLoggedIn={true} />
            <Button variant="ghost" size="sm" render={<Link to={`post/${message.postUid}`} />}>
              jump to post
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
