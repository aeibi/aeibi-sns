import { Button } from "@/components/ui/button"
import { cn, formatCount } from "@/lib/utils"

type PostActionButtonProps = React.ComponentProps<typeof Button> & {
  icon: React.ReactNode
  count: number
  active?: boolean
}

export function PostActionButton({ icon, count, active = false, onClick, ...props }: PostActionButtonProps) {
  return (
    <Button
      variant="ghost"
      size="sm"
      onClick={onClick}
      aria-pressed={active}
      className={cn("text-muted-foreground hover:text-foreground", active && "text-primary hover:text-primary")}
      {...props}
    >
      {icon}
      <span>{formatCount(count)}</span>
    </Button>
  )
}
