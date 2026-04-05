import { Button } from "@/components/ui/button"
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card"

export type MessageCategory = "follow" | "comment"

type MessageCategorySidenavProps = React.ComponentProps<typeof Card> & {
  selectedCategory: MessageCategory
  onCategoryChange: (category: MessageCategory) => void
}

export function MessageCategorySidenav({ selectedCategory, onCategoryChange, ...props }: MessageCategorySidenavProps) {
  return (
    <Card {...props}>
      <CardHeader>
        <CardTitle>Message Center</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-2">
        <Button size="sm" variant={selectedCategory === "follow" ? "default" : "outline"} onClick={() => onCategoryChange("follow")}>
          <span>Follow</span>
        </Button>
        <Button size="sm" variant={selectedCategory === "comment" ? "default" : "outline"} onClick={() => onCategoryChange("comment")}>
          <span>Comment</span>
        </Button>
      </CardContent>
    </Card>
  )
}
