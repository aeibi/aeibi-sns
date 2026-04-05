import { Image } from "@unpic/react"
import { XIcon } from "lucide-react"
import ImageViewer from "@/components/image-viewer"
import { Button } from "@/components/ui/button"

type PostComposerMediaProps = {
  images: string[]
  onRemove: (index: number) => void
}

export function PostComposerMedia({ images, onRemove }: PostComposerMediaProps) {
  if (!images.length) return null

  return (
    <div className="flex flex-wrap gap-2">
      {images.map((image, index) => (
        <div key={`${image}-${index}`} className="relative h-36 w-36">
          <ImageViewer
            images={images}
            initialIndex={index}
            trigger={<button type="button" className="h-full w-full overflow-hidden rounded-md" />}
          >
            <Image
              src={image}
              alt={`Image ${index + 1}`}
              layout="fixed"
              width={144}
              height={144}
              className="h-full w-full cursor-zoom-in object-cover"
            />
          </ImageViewer>
          <Button
            variant="secondary"
            size="icon-xs"
            className="absolute top-1 right-1"
            aria-label={`Remove image ${index + 1}`}
            onClick={() => onRemove(index)}
          >
            <XIcon />
          </Button>
        </div>
      ))}
    </div>
  )
}
