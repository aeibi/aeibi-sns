"use client"

import { Button } from "@/components/ui/button"
import { useSidebar } from "@/components/ui/sidebar"
import { MessageCircleIcon, PanelLeftIcon, StarIcon } from "lucide-react"
import { Separator } from "@/components/ui/separator"
import { HeaderNav } from "@/components/header-nav"
import { HeaderAction } from "@/components/header-action"
import { siGithub } from "simple-icons"
import { Link } from "react-router-dom"
import { SearchForm } from "@/components/search-form"
import { HeaderUser } from "@/components/header-user"
import type { User } from "@/types/user"

const data = {
  nav: [
    {
      name: "HOME",
      url: "/",
    },
  ],
  action: [
    {
      icon: <StarIcon />,
      url: "/favorites",
    },
    {
      icon: <MessageCircleIcon />,
      url: "/messages",
    },
  ],
}

export function SiteHeader({ user }: { user?: User }) {
  const { toggleSidebar } = useSidebar()
  return (
    <header className="sticky top-0 z-50 flex w-full items-center border-b bg-background">
      <div className="flex h-(--header-height) w-full items-center gap-2 px-4">
        <Button variant="ghost" size="icon" onClick={toggleSidebar}>
          <PanelLeftIcon />
        </Button>
        <Separator orientation="vertical" className="data-vertical:h-4 data-vertical:self-auto" />
        <HeaderNav items={data.nav} />
        <div className="w-full" />
        <SearchForm />
        <Separator orientation="vertical" className="data-vertical:h-4 data-vertical:self-auto" />
        <HeaderAction items={data.action} />
        <Separator orientation="vertical" className="data-vertical:h-4 data-vertical:self-auto" />
        <Button variant="ghost" size="icon-lg" render={<Link to="https://github.com/aeibi" />}>
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path d={siGithub.path} fill="currentColor" />
          </svg>
        </Button>
        <Separator orientation="vertical" className="data-vertical:h-4 data-vertical:self-auto" />
        {user ? (
          <HeaderUser user={user} />
        ) : (
          <Button variant="outline" size="sm" render={<Link to="/login" />}>
            Login
          </Button>
        )}
      </div>
    </header>
  )
}
