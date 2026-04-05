import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { UserDropdownMenu } from "@/components/user-dropdown-menu"
import type { User } from "@/types/user"

export function HeaderUser({ user }: { user: User }) {
  return (
    <UserDropdownMenu user={user} trigger={<Button variant="ghost" size="icon-lg" className="rounded-full" />}>
      <Avatar>
        <AvatarImage src={user.avatarUrl} alt={user.uid} />
        <AvatarFallback />
      </Avatar>
    </UserDropdownMenu>
  )
}
