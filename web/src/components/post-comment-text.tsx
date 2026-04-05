import { useLayoutEffect, useRef, useState } from "react"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"

export function PostCommentText({ text }: { text: string }) {
  const ref = useRef<HTMLParagraphElement>(null)
  const [clamped, setClamped] = useState(false)
  const [expanded, setExpanded] = useState(false)

  useLayoutEffect(() => {
    const el = ref.current
    if (el) setClamped(el.scrollHeight > el.clientHeight)
  }, [text])

  return (
    <div>
      <p ref={ref} className={cn("text-sm whitespace-pre-wrap", !expanded && "line-clamp-10")}>
        {text}
      </p>
      {clamped && !expanded && (
        <Button variant="link" size="sm" className="h-auto p-0" onClick={() => setExpanded(true)}>
          More
        </Button>
      )}
    </div>
  )
}
