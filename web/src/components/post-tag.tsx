import { Badge } from "@/components/ui/badge"
import { useNavigate } from "react-router-dom"

export function PostTag({ tags }: { tags: string[] }) {
  const navigate = useNavigate()
  if (!tags.length) return null
  return (
    <div className="flex flex-wrap gap-2">
      {tags.map((tag) => (
        <Badge key={tag} variant="outline" onClick={() => navigate(`/tag?tag=${tag}`)}>
          <span className="cursor-pointer truncate hover:underline"># {tag}</span>
        </Badge>
      ))}
    </div>
  )
}
