"use client"

import { NavMain } from "@/components/sidebar-nav-main"
import { NavSecondary } from "@/components/sidebar-nav-secondary"
import { NavUser } from "@/components/sidebar-nav-user"
import { Sidebar, SidebarContent, SidebarFooter } from "@/components/ui/sidebar"
import {
  SendIcon,
  BookOpenIcon,
  HouseIcon,
  MessageCircleIcon,
  Settings2Icon,
  StarIcon,
  UserIcon,
  UserRoundPlusIcon,
  UsersIcon,
} from "lucide-react"
import { siGithub } from "simple-icons"
import type { User } from "@/types/user"
import { Button } from "@/components/ui/button"
import { Link } from "react-router-dom"

const data = {
  navSecondary: [
    {
      title: "GitHub",
      url: "https://github.com/aeibi",
      icon: (
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <path d={siGithub.path} fill="currentColor" />
        </svg>
      ),
    },
    {
      title: "Feedback",
      url: "https://github.com/aeibi/aeibi-api/issues/new/choose",
      icon: <SendIcon />,
    },
    {
      title: "Help",
      url: "https://github.com/aeibi/aeibi-api/issues",
      icon: <BookOpenIcon />,
    },
  ],
  navMain: [
    {
      name: "Home",
      url: "/",
      icon: <HouseIcon />,
    },
    {
      name: "Favorites",
      url: "/favorites",
      icon: <StarIcon />,
    },
    {
      name: "Messages",
      url: "/messages",
      icon: <MessageCircleIcon />,
    },
    {
      name: "Profile",
      url: "/profile",
      icon: <UserIcon />,
    },
    {
      name: "Following",
      url: "/relation?tab=following",
      icon: <UserRoundPlusIcon />,
    },
    {
      name: "Followers",
      url: "/relation?tab=followers",
      icon: <UsersIcon />,
    },
    {
      name: "Profile Center",
      url: "/profile-center",
      icon: <Settings2Icon />,
    },
  ],
}
export function AppSidebar({ user }: { user?: User }) {
  return (
    <Sidebar className="top-(--header-height) h-[calc(100svh-var(--header-height))]!">
      <SidebarContent>
        <NavMain items={data.navMain} />
        <NavSecondary items={data.navSecondary} className="mt-auto" />
      </SidebarContent>
      <SidebarFooter>
        {user ? (
          <NavUser user={user} />
        ) : (
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" className="flex-1" render={<Link to="/login" />}>
              Login
            </Button>
            <Button variant="outline" size="sm" className="flex-1" render={<Link to="/signup" />}>
              Sign up
            </Button>
          </div>
        )}
      </SidebarFooter>
    </Sidebar>
  )
}
