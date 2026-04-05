import { Link } from "react-router-dom"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Card, CardContent } from "@/components/ui/card"
import { formatCount } from "@/lib/utils"
import type { User } from "@/types/user"
import type { RelationCategory } from "@/components/relation-category-sidenav"

type RelationUserCardProps = {
  user: User
  relation: RelationCategory
}

export function RelationUserCard({ user }: RelationUserCardProps) {
  return (
    <Card>
      <CardContent>
        <Link to={`/profile?uid=${encodeURIComponent(user.uid)}`} className="group flex items-start gap-3">
          <Avatar size="lg">
            <AvatarImage src={user.avatarUrl} alt={user.nickname} />
            <AvatarFallback />
          </Avatar>
          <div className="flex flex-1 flex-col gap-2">
            <div className="flex flex-wrap items-center gap-2">
              <p className="text-sm font-semibold group-hover:underline">{user.nickname}</p>
            </div>
            {user.description && <p className="text-sm wrap-break-word text-muted-foreground">{user.description}</p>}
            <div className="flex items-center gap-3 text-xs text-muted-foreground">
              <span>Following {formatCount(user.followingCount)}</span>
              <span>Followers {formatCount(user.followersCount)}</span>
            </div>
          </div>
        </Link>
      </CardContent>
    </Card>
  )
}
