"use client"

import { SidebarGroup, SidebarMenu, SidebarMenuButton, SidebarMenuItem } from "@/components/ui/sidebar"
import { Link, useLocation } from "react-router-dom"

export function NavMain({
  items,
}: {
  items: {
    name: string
    url: string
    icon: React.ReactNode
  }[]
}) {
  const { pathname, search } = useLocation()
  return (
    <SidebarGroup>
      <SidebarMenu className="gap-2">
        {items.map((item) => (
          <SidebarMenuItem key={item.name}>
            <SidebarMenuButton isActive={isMenuItemActive(item.url, pathname, search)} render={<Link to={item.url} />}>
              {item.icon}
              <span>{item.name}</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        ))}
      </SidebarMenu>
    </SidebarGroup>
  )
}

function isMenuItemActive(target: string, pathname: string, search: string) {
  const [targetPath, targetQuery] = target.split("?")
  if (pathname !== targetPath) return false
  if (!targetQuery) return true

  const currentParams = new URLSearchParams(search)
  const targetParams = new URLSearchParams(targetQuery)
  return Array.from(targetParams.entries()).every(([key, value]) => currentParams.get(key) === value)
}
