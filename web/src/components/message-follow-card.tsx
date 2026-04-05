import { Avatar, AvatarBadge, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Card, CardContent } from "@/components/ui/card"
import { formatDateTime } from "@/lib/utils"
import type { FollowMessage } from "@/types/message"
import { Link } from "react-router-dom"

type MessageFollowCardProps = {
  message: FollowMessage
}

export function MessageFollowCard({ message }: MessageFollowCardProps) {
  const action = "followed you."
  return (
    <Card>
      <CardContent>
        <Link to={`/profile?uid=${encodeURIComponent(message.actor.uid)}`} className="group flex gap-3">
          <Avatar size="lg">
            <AvatarImage src={message.actor.avatarUrl} alt={message.actor.nickname} />
            <AvatarFallback />
            {!message.isRead && <AvatarBadge className="bg-red-300" />}
          </Avatar>
          <div className="flex-1">
            <p className="flex gap-2 text-sm">
              <span className="font-semibold group-hover:underline">{message.actor.nickname}</span>
              <span>{action}</span>
            </p>
            <span className="text-xs text-muted-foreground">{formatDateTime(message.createdAt)}</span>
          </div>
        </Link>
      </CardContent>
    </Card>
  )
}
