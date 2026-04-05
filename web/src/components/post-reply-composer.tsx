import { useCommentServiceCreateReply } from "@/api/generated"
import { useState, type SubmitEvent } from "react"
import { Dialog, DialogContent, DialogTrigger } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupTextarea } from "@/components/ui/input-group"
import { toast } from "sonner"

export function PostReplyComposer({
  parentUid,
  isLoggedIn,
  onPosted,
}: {
  parentUid: string
  isLoggedIn: boolean
  onPosted?: (uid: string) => void
}) {
  const { mutate, isPending } = useCommentServiceCreateReply()
  const [open, setOpen] = useState(false)
  const showLoginToast = () => {
    toast.error("Please log in first.", { position: "top-center" })
  }
  const handleOpenChange = (nextOpen: boolean) => {
    if (nextOpen && !isLoggedIn) {
      showLoginToast()
      return
    }
    setOpen(nextOpen)
  }
  const handleSubmit = (event: SubmitEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (!isLoggedIn) {
      showLoginToast()
      return
    }
    const form = event.currentTarget
    const content = new FormData(form).get("content")?.toString().trim()
    if (!content) return
    mutate(
      { parentUid, data: { parentUid, content } },
      {
        onSuccess: (data) => {
          onPosted?.(data.uid)
          setOpen(false)
        },
      }
    )
  }
  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger render={<Button type="button" variant="ghost" size="xs" className="text-muted-foreground hover:text-foreground" />}>
        reply
      </DialogTrigger>
      <DialogContent showCloseButton={false} className="sm:max-w-lg">
        <form className={"flex min-w-0 items-start gap-2"} onSubmit={handleSubmit}>
          <InputGroup className="min-w-0">
            <InputGroupTextarea
              id="block-end-textarea"
              name="content"
              placeholder="Write a comment..."
              rows={1}
              className="min-h-0 wrap-break-word"
              required
            />
            <InputGroupAddon align="block-end">
              <InputGroupButton type="submit" variant="default" size="xs" className="ml-auto" disabled={isPending}>
                {isPending ? "Posting..." : "Post"}
              </InputGroupButton>
            </InputGroupAddon>
          </InputGroup>
        </form>
      </DialogContent>
    </Dialog>
  )
}
