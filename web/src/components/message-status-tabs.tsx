import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"

export type MessageStatus = "unread" | "all"

interface MessageStatusTabsProps {
  selectedStatus: MessageStatus
  onStatusChange: (status: MessageStatus) => void
  onMarkAllAsRead: () => void
}

export function MessageStatusTabs({ selectedStatus, onStatusChange, onMarkAllAsRead }: MessageStatusTabsProps) {
  return (
    <Card>
      <CardContent className="flex gap-2">
        <Button size="sm" variant={selectedStatus === "unread" ? "default" : "ghost"} onClick={() => onStatusChange("unread")}>
          unread
        </Button>
        <Button size="sm" variant={selectedStatus === "all" ? "default" : "ghost"} onClick={() => onStatusChange("all")}>
          all
        </Button>
        <Button size="sm" variant="outline" className="ml-auto" onClick={onMarkAllAsRead}>
          Mark all as read
        </Button>
      </CardContent>
    </Card>
  )
}
