import { Card, CardContent } from "@/components/ui/card"
import { Textarea } from "@/components/ui/textarea"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { HashIcon, ImagesIcon } from "lucide-react"
import { useMemo, useState, type ChangeEvent, type SubmitEvent } from "react"
import { usePostServiceCreatePost } from "@/api/generated"
import { Label } from "@/components/ui/label"
import { useUploadFiles } from "@/hooks/use-upload-files"
import { toast } from "sonner"
import { PostComposerTag } from "@/components/post-composer-tag"
import { PostComposerMedia } from "@/components/post-composer-media"

type PostComposerProps = React.ComponentProps<typeof Card> & {
  onPosted: (uid: string) => void
}

export function PostComposerCard({ onPosted, ...props }: PostComposerProps) {
  const [tagsInput, setTagsInput] = useState("")
  const tags = useMemo(() => Array.from(new Set(tagsInput.trim().split(/\s+/).filter(Boolean))), [tagsInput])
  const [images, setImages] = useState<string[]>([])

  const { mutate: createPost } = usePostServiceCreatePost()
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
    const text = formData.get("text")?.toString().trim()
    if (!text) return
    createPost(
      { data: { text, images, tags } },
      {
        onSuccess: (data) => {
          form.reset()
          setTagsInput("")
          setImages([])
          onPosted(data.uid)
        },
        onError: () => {
          toast.error("Failed to create the post.", { position: "top-center" })
        },
      }
    )
  }

  return (
    <Card {...props}>
      <CardContent>
        <form className="flex flex-col gap-3" onSubmit={handleSubmit}>
          <Textarea name="text" placeholder="What's happening?" required />
          <Input
            id="tag-input"
            name="tag-input"
            placeholder="tags, split by space"
            value={tagsInput}
            onChange={(event) =>
              setTagsInput(
                event.currentTarget.value
                  .replace(/[^\p{L}\p{N}\s]/gu, "")
                  .replace(/\s+/g, " ")
                  .trimStart()
              )
            }
            spellCheck={false}
            autoComplete="off"
            className="not-focus:absolute not-focus:h-0 not-focus:border-0 not-focus:py-0"
          />
          <PostComposerTag tags={tags} />
          <input
            id="image-input"
            type="file"
            accept="image/*"
            multiple
            disabled={isPending}
            className="hidden"
            onChange={handleImageUpload}
          />
          <PostComposerMedia
            images={images}
            onRemove={(index) => setImages((current) => current.filter((_, currentIndex) => currentIndex !== index))}
          />
          <div className="flex">
            <Button type="button" variant="ghost" render={<Label htmlFor="tag-input" />}>
              <HashIcon />
              tag
            </Button>
            <Button type="button" variant="ghost" disabled={isPending} render={<Label htmlFor="image-input" />}>
              <ImagesIcon />
              {isPending ? "uploading..." : "image"}
            </Button>
            <Button type="submit" className="ml-auto">
              post
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}
