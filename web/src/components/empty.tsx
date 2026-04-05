import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyTitle } from "@/components/ui/empty"
import { ArrowLeftIcon, RefreshCcwIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { useNavigate } from "react-router-dom"

export function EmptyState() {
  const navigate = useNavigate()
  return (
    <Empty className="h-full bg-muted">
      <EmptyHeader>
        <EmptyTitle>Nothing to show</EmptyTitle>
        <EmptyDescription className="max-w-xs text-pretty">There is no content to display right now.</EmptyDescription>
      </EmptyHeader>
      <EmptyContent className="flex">
        <Button variant="outline" onClick={() => navigate(-1)}>
          <ArrowLeftIcon data-icon="inline-start" />
          Back
        </Button>
        <Button variant="outline" onClick={() => navigate(0)}>
          <RefreshCcwIcon data-icon="inline-start" />
          Refresh
        </Button>
      </EmptyContent>
    </Empty>
  )
}
