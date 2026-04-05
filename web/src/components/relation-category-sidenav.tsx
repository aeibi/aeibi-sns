import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

export type RelationCategory = "following" | "followers"

type RelationCategorySidenavProps = React.ComponentProps<typeof Card> & {
  selectedCategory: RelationCategory
  onCategoryChange: (category: RelationCategory) => void
}

export function RelationCategorySidenav({ selectedCategory, onCategoryChange, ...props }: RelationCategorySidenavProps) {
  return (
    <Card {...props}>
      <CardHeader>
        <CardTitle>Relations</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-2">
        <Button size="sm" variant={selectedCategory === "following" ? "default" : "outline"} onClick={() => onCategoryChange("following")}>
          <span>Following</span>
        </Button>
        <Button size="sm" variant={selectedCategory === "followers" ? "default" : "outline"} onClick={() => onCategoryChange("followers")}>
          <span>Followers</span>
        </Button>
      </CardContent>
    </Card>
  )
}
