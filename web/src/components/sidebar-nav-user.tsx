"use client"

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { SidebarMenu, SidebarMenuButton, SidebarMenuItem, useSidebar } from "@/components/ui/sidebar"
import type { User } from "@/types/user"
import { ChevronsUpDownIcon } from "lucide-react"
import { UserDropdownMenu } from "@/components/user-dropdown-menu"

export function NavUser({ user }: { user: User }) {
  const { isMobile } = useSidebar()
  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <UserDropdownMenu
          user={user}
          trigger={<SidebarMenuButton size="lg" className="aria-expanded:bg-muted aria-expanded:text-foreground" />}
          side={isMobile ? "bottom" : "right"}
        >
          <Avatar>
            <AvatarImage src={user.avatarUrl} alt={user.uid} />
            <AvatarFallback />
          </Avatar>
          <div className="grid flex-1 text-left text-sm leading-tight">
            <span className="truncate font-medium">{user.nickname}</span>
            <span className="truncate text-xs">{user.email}</span>
          </div>
          <ChevronsUpDownIcon className="ml-auto size-4" />
        </UserDropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  )
}
