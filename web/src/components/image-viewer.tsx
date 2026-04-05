"use client"

import { cn } from "@/lib/utils"
import { Dialog, DialogContent, DialogTrigger } from "@/components/ui/dialog"
import { ChevronLeftIcon, ChevronRightIcon, Download, MinusCircle, PlusCircle } from "lucide-react"
import { useState } from "react"
import { Button } from "@/components/ui/button"
import { TransformComponent, TransformWrapper } from "react-zoom-pan-pinch"
import { Image } from "@unpic/react"

type ImageViewerProps = React.ComponentProps<typeof DialogContent> & {
  trigger: React.ComponentProps<typeof DialogTrigger>["render"]
  images: string[]
  initialIndex: number
}

function ImageViewer({ images, initialIndex, trigger, children, className, ...props }: ImageViewerProps) {
  const [activeIndex, setActiveIndex] = useState(initialIndex)
  const activeImage = images[activeIndex]

  const handleDownload = () => {
    const link = document.createElement("a")
    link.href = activeImage
    link.download = activeImage.split("/").pop() ?? "image"
    link.target = "_blank"
    link.rel = "noopener noreferrer"
    link.click()
  }

  return (
    <Dialog onOpenChange={() => setActiveIndex(initialIndex)}>
      <DialogTrigger render={trigger}>{children}</DialogTrigger>
      <DialogContent
        showCloseButton={false}
        className={cn("h-[95vh] w-[95vw] max-w-none overflow-hidden p-0 sm:max-w-none", className)}
        {...props}
      >
        <div className="relative flex items-center justify-center bg-background">
          <TransformWrapper initialScale={1} initialPositionX={0} initialPositionY={0}>
            {({ zoomIn, zoomOut }) => (
              <>
                <TransformComponent>
                  <Image
                    src={activeImage}
                    alt="Image - Full"
                    layout="fullWidth"
                    objectFit="contain"
                    className="h-[95vh] w-[95vw] object-contain"
                  />
                </TransformComponent>
                <div className="absolute bottom-4 left-1/2 z-10 flex -translate-x-1/2 gap-2">
                  <Button variant="outline" size="icon-lg" onClick={() => zoomOut()} className="cursor-pointer" aria-label="Zoom out">
                    <MinusCircle />
                  </Button>
                  <Button variant="outline" size="icon-lg" onClick={() => zoomIn()} className="cursor-pointer" aria-label="Zoom in">
                    <PlusCircle />
                  </Button>
                  <Button variant="outline" size="icon-lg" onClick={handleDownload} className="cursor-pointer" aria-label="Download image">
                    <Download />
                  </Button>
                </div>
              </>
            )}
          </TransformWrapper>
          {activeIndex > 0 && (
            <Button
              type="button"
              variant="secondary"
              size="icon-lg"
              aria-label="Previous image"
              className="absolute left-2"
              onClick={() => setActiveIndex((current) => current - 1)}
            >
              <ChevronLeftIcon />
            </Button>
          )}
          {activeIndex < images.length - 1 && (
            <Button
              type="button"
              variant="secondary"
              size="icon-lg"
              aria-label="Next image"
              className="absolute right-2"
              onClick={() => setActiveIndex((current) => current + 1)}
            >
              <ChevronRightIcon />
            </Button>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}

export default ImageViewer
