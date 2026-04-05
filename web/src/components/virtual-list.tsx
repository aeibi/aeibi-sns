import { type ReactNode, useEffect, useRef } from "react"
import { defaultRangeExtractor, useVirtualizer } from "@tanstack/react-virtual"

type VirtualListProps<T> = {
  items: T[]
  header?: ReactNode
  stickyHeader?: ReactNode

  renderItem: (item: T, index: number) => ReactNode
  getItemKey?: (item: T, index: number) => string | number

  hasNextPage?: boolean
  isFetchingNextPage?: boolean
  onLoadMore?: () => Promise<unknown> | unknown

  estimateSize?: (index: number) => number
  gap?: number
  paddingStart?: number
  paddingEnd?: number

  className?: string
  innerClassName?: string
}

export function VirtualList<T>({
  items,
  header = <></>,
  stickyHeader,
  renderItem,
  getItemKey,
  hasNextPage = false,
  isFetchingNextPage = false,
  onLoadMore,
  estimateSize = () => 600,
  gap = 16,
  paddingStart = 16,
  paddingEnd = 16,
  className = "h-full w-full overflow-y-auto",
  innerClassName = "mx-auto w-full max-w-4xl px-4",
}: VirtualListProps<T>) {
  const scrollRef = useRef<HTMLDivElement>(null)

  const count = items.length + 1

  const toDataIndex = (virtualIndex: number) => virtualIndex - 1

  const virtualizer = useVirtualizer({
    count,
    estimateSize,
    getScrollElement: () => scrollRef.current,
    getItemKey: (index) => {
      if (index === 0) return "__header__"

      const dataIndex = toDataIndex(index)
      const item = items[dataIndex]

      if (!item) return index
      return getItemKey ? getItemKey(item, dataIndex) : dataIndex
    },
    rangeExtractor: (range) => {
      const next = new Set([0, ...defaultRangeExtractor(range)])
      return [...next].sort((a, b) => a - b)
    },
    gap,
    paddingStart,
    paddingEnd,
  })

  useEffect(() => {
    virtualizer.shouldAdjustScrollPositionOnItemSizeChange = (item, _delta, instance) => {
      const scrollOffset = instance.scrollOffset ?? 0
      return item.end <= scrollOffset
    }
  }, [virtualizer])

  const virtualItems = virtualizer.getVirtualItems()

  useEffect(() => {
    if (!onLoadMore || !hasNextPage || isFetchingNextPage) return
    if (items.length === 0) return

    const lastDataVirtualItem = [...virtualItems].reverse().find((item) => item.index >= 1)

    if (!lastDataVirtualItem) return

    const lastDataIndex = toDataIndex(lastDataVirtualItem.index)
    if (lastDataIndex < items.length - 1) return

    void onLoadMore()
  }, [hasNextPage, isFetchingNextPage, items.length, onLoadMore, virtualItems])

  return (
    <div ref={scrollRef} className={className}>
      <div className="relative" style={{ height: `${virtualizer.getTotalSize()}px` }}>
        {stickyHeader}
        {virtualItems.map((virtualItem) => {
          const isHeader = virtualItem.index === 0
          const dataIndex = toDataIndex(virtualItem.index)
          const item = items[dataIndex]

          return (
            <div
              key={virtualItem.key}
              ref={virtualizer.measureElement}
              data-index={virtualItem.index}
              className="absolute top-0 left-0 w-full"
              style={{ transform: `translateY(${virtualItem.start}px)` }}
            >
              <div className={innerClassName}>{isHeader ? header : item != null ? renderItem(item, dataIndex) : null}</div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
