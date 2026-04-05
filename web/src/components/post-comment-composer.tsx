import type { User } from "@/types/user"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupTextarea } from "@/components/ui/input-group"
import { ImagesIcon } from "lucide-react"
import { useState, type ChangeEvent, type SubmitEvent } from "react"
import { useUploadFiles } from "@/hooks/use-upload-files"
import { useCommentServiceCreateTopComment } from "@/api/generated"
import { toast } from "sonner"
import { PostComposerMedia } from "@/components/post-composer-media"
import { Label } from "@/components/ui/label"
import { cn } from "@/lib/utils"

type PostCommentsComposerProps = React.ComponentProps<"form"> & {
  user: User
  postUid: string
  onPosted?: (uid: string) => void
}

export function PostCommentsComposer({ user, postUid, onPosted, className, ...props }: PostCommentsComposerProps) {
  const [images, setImages] = useState<string[]>([])
  const { mutate: createComment } = useCommentServiceCreateTopComment()
  const { mutate: uploadFiles, isPending } = useUploadFiles()
  const handleImageUpload = async (event: ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(event.target.files ?? []).filter((file) => file.type.startsWith("image/"))
    event.currentTarget.value = ""
    if (!files.length) return
    uploadFiles(
      {
        files,
        onFileUploaded: ({ response }) => {
          setImages((current) => Array.from(new Set([...current, response.url])))
        },
      },
      {
        onError: () => {
          toast.error("Failed to upload images", { position: "top-center" })
        },
      }
    )
  }

  const handleSubmit = (event: SubmitEvent<HTMLFormElement>) => {
    event.preventDefault()
    const form = event.currentTarget
    const formData = new FormData(form)
    const content = formData.get("content")?.toString().trim()
    if (!content) return
    createComment(
      {
        postUid,
        data: { postUid, content, images },
      },
      {
        onSuccess: (data) => {
          form.reset()
          setImages([])
          onPosted?.(data.uid)
        },
      }
    )
  }
  return (
    <form className={cn("flex items-start gap-2", className)} onSubmit={handleSubmit} {...props}>
      <Avatar>
        <AvatarImage src={user.avatarUrl} alt={user.nickname} />
        <AvatarFallback />
      </Avatar>
      <div className="flex min-w-0 flex-1 flex-col gap-2">
        <InputGroup>
          <InputGroupTextarea name="content" placeholder="Write a comment..." rows={1} className="min-h-0 wrap-break-word" required />
          <InputGroupAddon
            align="block-end"
            className="hidden group-focus-within/input-group:flex group-has-[[data-slot=input-group-control]:not(:placeholder-shown)]/input-group:flex"
          >
            <input
              id={`${postUid}-image-input`}
              type="file"
              accept="image/*"
              multiple
              disabled={isPending}
              className="hidden"
              onChange={handleImageUpload}
            />
            <InputGroupButton variant="ghost" render={<Label htmlFor={`${postUid}-image-input`} />}>
              <ImagesIcon />
            </InputGroupButton>
            <InputGroupButton type="submit" variant="default" className="ml-auto">
              post
            </InputGroupButton>
          </InputGroupAddon>
        </InputGroup>
        <PostComposerMedia
          images={images}
          onRemove={(index) => setImages((current) => current.filter((_, currentIndex) => currentIndex !== index))}
        />
      </div>
    </form>
  )
}
