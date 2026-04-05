import { cn } from "@/lib/utils"
import type { Post } from "@/types/post"
import { Image } from "@unpic/react"
import ImageViewer from "@/components/image-viewer"

export function PostMedia({ post }: { post: Post }) {
  const count = post.images.length
  if (!count) return null
  return (
    <div
      className={cn("@container grid w-full gap-2", {
        "grid-cols-1 justify-items-start": count === 1,
        "grid-cols-2": count === 2 || count === 4,
        "grid-cols-3": count === 3 || count >= 5,
      })}
    >
      {post.images.slice(0, 9).map((image, index) => (
        <ImageViewer
          key={`${image}-${index}`}
          images={post.images}
          initialIndex={index}
          trigger={<button type="button" className="cursor-zoom-in overflow-hidden rounded-xl" />}
        >
          <div className="relative">
            {count === 1 ? (
              <Image
                src={image}
                alt={`Post image ${index + 1}`}
                layout="fullWidth"
                loading="lazy"
                className="h-auto max-h-[100cqw] w-auto max-w-full object-contain"
              />
            ) : (
              <Image
                src={image}
                alt={`Post image ${index + 1}`}
                layout="fullWidth"
                loading="lazy"
                className="aspect-square w-full object-cover object-center"
              />
            )}
            {index === 8 && count > 9 && (
              <div className="absolute inset-0 grid place-items-center bg-background/70 text-xl font-semibold">+{count - 9}</div>
            )}
          </div>
        </ImageViewer>
      ))}
    </div>
  )
}
