import { Badge } from "@/components/ui/badge"

type PostComposerTagProps = {
  tags: string[]
}

export function PostComposerTag({ tags }: PostComposerTagProps) {
  if (!tags.length) return null
  return (
    <div className="flex flex-wrap gap-2">
      {tags.map((tag) => (
        <Badge key={tag} variant="outline">
          # {tag}
        </Badge>
      ))}
    </div>
  )
}
