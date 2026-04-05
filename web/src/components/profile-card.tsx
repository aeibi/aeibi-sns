import { Button } from "@/components/ui/button"
import { cn, formatCount } from "@/lib/utils"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import type { User } from "@/types/user"
import { FlagIcon, PencilIcon, Share2Icon } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { Link } from "react-router-dom"
import { toast } from "sonner"
import { useCopyToClipboard } from "@/hooks/use-copy-to-clipboard"

type ProfileCardProps = React.ComponentProps<typeof Card> & {
  user: User
  me?: User
  onFollow?: () => void
  followPending?: boolean
}

export function ProfileCard({ className, user, me, onFollow, followPending = false, ...props }: ProfileCardProps) {
  const isOwnProfile = !!me && me.uid === user.uid

  const { copy } = useCopyToClipboard()
  const handleCopy = async () => {
    const ok = await copy(`${window.location.origin}/profile?uid=${encodeURIComponent(user.uid)}`)
    if (ok) toast.success("Share link copied to clipboard.", { position: "top-center" })
    else toast.error("Failed to copy the share link.", { position: "top-center" })
  }

  return (
    <Card className={cn("", className)} {...props}>
      <CardHeader>
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="flex items-start gap-3">
            <Avatar className="size-20 shrink-0 bg-card ring-1 ring-border">
              <AvatarImage src={user.avatarUrl} alt={user.uid} />
              <AvatarFallback />
            </Avatar>
            <div className="flex min-w-0 flex-col gap-2 pt-1">
              <div className="flex flex-wrap items-center gap-2">
                <CardTitle className="text-xl tracking-tight">{user.nickname}</CardTitle>
                <Badge variant="outline">{user.role}</Badge>
              </div>
            </div>
          </div>
          <div className="flex flex-wrap items-center gap-2 sm:justify-end">
            <Button type="button" variant="outline" size="sm" onClick={handleCopy}>
              <Share2Icon />
              <span>Share Profile</span>
            </Button>
            {!!me && !isOwnProfile && (
              <Button
                type="button"
                variant="outline"
                size="sm"
                aria-label="Follow user"
                disabled={!onFollow || followPending}
                onClick={onFollow}
              >
                {user.isFollowing ? "Following" : "Follow"}
              </Button>
            )}
            {!isOwnProfile && (
              <Button type="button" variant="outline" size="icon-sm" aria-label="Report user">
                <FlagIcon />
              </Button>
            )}
            {isOwnProfile && (
              <Button type="button" variant="outline" size="sm" render={<Link to="/profile-center" />}>
                <PencilIcon />
                <span>Edit Profile</span>
              </Button>
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent className="flex flex-col">
        <div className="grid grid-cols-2 gap-2">
          <Button
            variant="outline"
            className="h-auto w-full flex-col items-center gap-0 px-3 py-2"
            render={isOwnProfile ? <Link to="/relation?tab=following" /> : undefined}
          >
            <p className="text-base font-semibold text-foreground">{formatCount(user.followingCount)}</p>
            <p className="text-xs text-muted-foreground">Following</p>
          </Button>
          <Button
            variant="outline"
            className="h-auto w-full flex-col items-center gap-0 px-3 py-2"
            render={isOwnProfile ? <Link to="/relation?tab=followers" /> : undefined}
          >
            <p className="text-base font-semibold text-foreground">{formatCount(user.followersCount)}</p>
            <p className="text-xs text-muted-foreground">Followers</p>
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
