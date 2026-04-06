import { Card, CardContent } from "@/components/ui/card"
import { Textarea } from "@/components/ui/textarea"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { HashIcon, ImagesIcon } from "lucide-react"
import { cn } from "@/lib/utils"
import { useMemo, useRef, useState, type ChangeEvent, type ClipboardEvent, type DragEvent, type SubmitEvent } from "react"
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
  const [isDragActive, setIsDragActive] = useState(false)
  const dragDepthRef = useRef(0)

  const { mutate: createPost } = usePostServiceCreatePost()
  const { mutate: uploadFiles, isPending } = useUploadFiles()

  const uploadImageFiles = (files: File[]) => {
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

  const handleImageUpload = (event: ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(event.target.files ?? []).filter((file) => file.type.startsWith("image/"))
    event.currentTarget.value = ""
    uploadImageFiles(files)
  }

  const handlePasteImage = (event: ClipboardEvent<HTMLTextAreaElement>) => {
    if (isPending) return
    const clipboardItems = Array.from(event.clipboardData.items)
    const itemFiles = clipboardItems
      .filter((item) => item.kind === "file" && item.type.startsWith("image/"))
      .map((item) => item.getAsFile())
      .filter((file): file is File => file !== null)
    const files = Array.from(event.clipboardData.files).filter((file) => file.type.startsWith("image/"))
    uploadImageFiles(itemFiles.length ? itemFiles : files)
  }

  const getDraggedImageFiles = (event: DragEvent<HTMLElement>) =>
    Array.from(event.dataTransfer.files).filter((file) => file.type.startsWith("image/"))

  const hasDraggedImage = (event: DragEvent<HTMLElement>) => {
    const itemHasImage = Array.from(event.dataTransfer.items).some(
      (item) => item.kind === "file" && item.type.startsWith("image/")
    )
    if (itemHasImage) return true
    return getDraggedImageFiles(event).length > 0
  }

  const handleDragEnter = (event: DragEvent<HTMLElement>) => {
    if (isPending || !hasDraggedImage(event)) return
    event.preventDefault()
    event.stopPropagation()
    dragDepthRef.current += 1
    setIsDragActive(true)
  }

  const handleDragOver = (event: DragEvent<HTMLElement>) => {
    if (isPending || !hasDraggedImage(event)) return
    event.preventDefault()
    event.stopPropagation()
    event.dataTransfer.dropEffect = "copy"
    if (!isDragActive) setIsDragActive(true)
  }

  const handleDragLeave = (event: DragEvent<HTMLElement>) => {
    if (!isDragActive) return
    event.preventDefault()
    event.stopPropagation()
    dragDepthRef.current = Math.max(0, dragDepthRef.current - 1)
    if (dragDepthRef.current === 0) setIsDragActive(false)
  }

  const handleDrop = (event: DragEvent<HTMLElement>) => {
    if (isPending || !hasDraggedImage(event)) return
    event.preventDefault()
    event.stopPropagation()
    dragDepthRef.current = 0
    setIsDragActive(false)
    uploadImageFiles(getDraggedImageFiles(event))
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
    <Card
      {...props}
      className={cn("transition-colors", isDragActive && "border-primary/60 bg-primary/5", props.className)}
      onDragEnter={handleDragEnter}
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
    >
      <CardContent>
        <form className="flex flex-col gap-3" onSubmit={handleSubmit}>
          {isDragActive && (
            <div className="rounded-lg border border-dashed border-primary/60 bg-primary/5 px-3 py-2 text-xs text-primary">
              Drop images here
            </div>
          )}
          <Textarea name="text" placeholder="What's happening?" onPaste={handlePasteImage} required />
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
