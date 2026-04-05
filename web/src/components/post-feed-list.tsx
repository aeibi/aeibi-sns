import { useEffect, useRef, type ReactNode } from "react"
import { defaultRangeExtractor, useVirtualizer } from "@tanstack/react-virtual"
import { PostCard } from "@/components/post-card"
import type { Post } from "@/types/post"
import type { User } from "@/types/user"

type PostFeedListProps = {
  posts: Post[]
  user: User | undefined
  headerKey?: string
  header?: ReactNode
  hasNextPage: boolean
  isFetchingNextPage: boolean
  onLoadMore: () => Promise<unknown>
  onUpdatePost: (uid: string, patch: Partial<Post>) => void
  onRemovePost: (uid: string) => void
}

export function PostFeedList({
  posts,
  user,
  headerKey,
  header,
  hasNextPage,
  isFetchingNextPage,
  onLoadMore,
  onUpdatePost,
  onRemovePost,
}: PostFeedListProps) {
  const ref = useRef<HTMLDivElement>(null)
  const hasHeader = header !== undefined
  const itemOffset = hasHeader ? 1 : 0

  // eslint-disable-next-line react-hooks/incompatible-library
  const virtualizer = useVirtualizer({
    count: posts.length + itemOffset,
    estimateSize: () => 600,
    getScrollElement: () => ref.current,
    getItemKey: (index) => {
      if (hasHeader && index === 0) return headerKey ?? "header"
      const postIndex = index - itemOffset
      return posts[postIndex]?.uid ?? index
    },
    rangeExtractor: (range) => {
      if (!hasHeader) return defaultRangeExtractor(range)
      const next = new Set([0, ...defaultRangeExtractor(range)])
      return [...next].sort((a, b) => a - b)
    },
    gap: 16,
    paddingStart: 16,
    paddingEnd: 16,
  })

  useEffect(() => {
    virtualizer.shouldAdjustScrollPositionOnItemSizeChange = (item, _delta, instance) => {
      const scrollOffset = instance.scrollOffset ?? 0
      return item.end <= scrollOffset
    }
  }, [virtualizer])

  const virtualItems = virtualizer.getVirtualItems()

  useEffect(() => {
    const lastPostItem = [...virtualItems].reverse().find((item) => item.index >= itemOffset)
    if (!lastPostItem) return
    if (!hasNextPage || isFetchingNextPage) return
    const lastPostIndex = lastPostItem.index - itemOffset
    if (lastPostIndex < posts.length - 1) return
    void onLoadMore()
  }, [hasNextPage, isFetchingNextPage, itemOffset, onLoadMore, posts.length, virtualItems])

  return (
    <div ref={ref} className="h-full w-full overflow-y-auto">
      <div className="relative" style={{ height: `${virtualizer.getTotalSize()}px` }}>
        {virtualItems.map((virtualItem) => {
          const isHeaderItem = hasHeader && virtualItem.index === 0
          const postIndex = virtualItem.index - itemOffset
          const post = postIndex >= 0 ? posts[postIndex] : undefined
          return (
            <div
              key={virtualItem.key}
              ref={virtualizer.measureElement}
              data-index={virtualItem.index}
              className="absolute top-0 left-0 w-full"
              style={{ transform: `translateY(${virtualItem.start}px)` }}
            >
              <div className="mx-auto w-full max-w-4xl px-4">
                {isHeaderItem
                  ? header
                  : !!post && (
                      <PostCard
                        user={user}
                        post={post}
                        onRemovePost={() => onRemovePost(post.uid)}
                        onUpdatePost={(patch) => onUpdatePost(post.uid, patch)}
                      />
                    )}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
