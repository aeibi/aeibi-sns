import ImageViewer from "@/components/image-viewer"
import { Image } from "@unpic/react"

type PostCommentMediaProps = React.ComponentProps<"div"> & {
  images: string[]
}

export function PostCommentMedia({ images }: PostCommentMediaProps) {
  if (!images.length) return null

  return (
    <div className="flex flex-wrap gap-2">
      {images.map((image, index) => (
        <div key={`${image}-${index}`} className="h-36 w-36">
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
        </div>
      ))}
    </div>
  )
}
