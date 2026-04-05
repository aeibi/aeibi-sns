import { BellIcon, LogOutIcon, Settings2Icon, UserIcon } from "lucide-react"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { cn } from "@/lib/utils"
import type { User } from "@/types/user"
import { Link } from "react-router-dom"
import { token } from "@/api/client"

type UserDropdownMenuProps = React.ComponentProps<typeof DropdownMenuContent> & {
  user: User
  trigger: React.ComponentProps<typeof DropdownMenuTrigger>["render"]
}

export function UserDropdownMenu({ user, trigger, children, className, ...props }: UserDropdownMenuProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger render={trigger}>{children}</DropdownMenuTrigger>
      <DropdownMenuContent side="bottom" align="end" sideOffset={4} className={cn("min-w-72 rounded-lg", className)} {...props}>
        <DropdownMenuGroup>
          <DropdownMenuLabel className="p-0 font-normal">
            <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
              <Avatar>
                <AvatarImage src={user.avatarUrl} alt={user.uid} />
                <AvatarFallback />
              </Avatar>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-medium text-foreground">{user.nickname}</span>
                <span className="truncate text-xs">{user.email}</span>
              </div>
            </div>
          </DropdownMenuLabel>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuGroup className="flex items-stretch gap-1">
          <DropdownMenuItem className="flex-1 flex-col items-center" render={<Link to="/relation?tab=following" />}>
            <span className="text-base leading-none font-semibold">{user.followingCount}</span>
            <span className="text-xs text-muted-foreground">Following</span>
          </DropdownMenuItem>
          <DropdownMenuItem className="flex-1 flex-col items-center" render={<Link to="/relation?tab=followers" />}>
            <span className="text-base leading-none font-semibold">{user.followersCount}</span>
            <span className="text-xs text-muted-foreground">Follower</span>
          </DropdownMenuItem>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          <DropdownMenuItem render={<Link to="/profile" />}>
            <UserIcon />
            Profile
          </DropdownMenuItem>
          <DropdownMenuItem render={<Link to="/profile-center" />}>
            <Settings2Icon />
            Profile Center
          </DropdownMenuItem>
          <DropdownMenuItem render={<Link to="/messages" />}>
            <BellIcon />
            Notifications
          </DropdownMenuItem>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          <DropdownMenuItem
            onClick={() => {
              token.clear()
              window.location.href = "/login"
            }}
          >
            <LogOutIcon />
            Log out
          </DropdownMenuItem>
        </DropdownMenuGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
