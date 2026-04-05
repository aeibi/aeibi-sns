import { Button } from "@/components/ui/button"
import { Link } from "react-router-dom"

export function HeaderNav({
  items,
}: {
  items: {
    name: string
    url: string
  }[]
}) {
  return (
    <>
      {items.map((item) => (
        <Button variant="ghost" size="lg" render={<Link to={item.url} />}>
          {item.name}
        </Button>
      ))}
    </>
  )
}
