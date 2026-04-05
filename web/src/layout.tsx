import { AppSidebar } from "@/components/app-sidebar"
import { SiteHeader } from "@/components/site-header"
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar"
import { Outlet } from "react-router-dom"
import { useUserServiceGetMe } from "./api/generated"

export function Layout() {
  const { data } = useUserServiceGetMe()
  return (
    <div className="h-svh overflow-hidden [--header-height:calc(--spacing(14))]">
      <SidebarProvider defaultOpen={false} className="flex h-full flex-col overflow-hidden">
        <SiteHeader user={data?.user} />
        <div className="flex min-h-0 flex-1">
          <AppSidebar user={data?.user} />
          <SidebarInset className="min-h-0 overflow-hidden">
            <Outlet />
          </SidebarInset>
        </div>
      </SidebarProvider>
    </div>
  )
}
