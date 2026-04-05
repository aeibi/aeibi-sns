import { useEffect, useRef } from "react"
import { useInfiniteQuery } from "@tanstack/react-query"
import { useVirtualizer } from "@tanstack/react-virtual"
import { useSearchParams } from "react-router-dom"
import {
  followServiceListMyFollowers,
  followServiceListMyFollowing,
  getFollowServiceListMyFollowersQueryKey,
  getFollowServiceListMyFollowingQueryKey,
  type FollowServiceListMyFollowersParams,
  type FollowServiceListMyFollowingParams,
} from "@/api/generated"
import { RelationCategorySidenav, type RelationCategory } from "@/components/relation-category-sidenav"
import { RelationSearchCard } from "@/components/relation-search-card"
import { RelationListSkeleton } from "@/components/loading-skeleton"
import { RelationUserCard } from "@/components/relation-user-card"
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/components/ui/empty"
import type { User } from "@/types/user"

export function Relation() {
  const [searchParams, setSearchParams] = useSearchParams()
  const category = searchParams.get("tab") === "followers" ? "followers" : "following"
  const query = searchParams.get("query") ?? ""

  const {
    data: followingData,
    fetchNextPage: fetchFollowingNextPage,
    isFetchingNextPage: isFetchingFollowingNextPage,
    hasNextPage: hasFollowingNextPage,
    isPending: isFollowingPending,
  } = useInfiniteQuery({
    queryKey: getFollowServiceListMyFollowingQueryKey({ query: query.trim() }),
    initialPageParam: { query: query.trim() } as FollowServiceListMyFollowingParams,
    queryFn: ({ pageParam, signal }) => followServiceListMyFollowing(pageParam, undefined, signal),
    getNextPageParam: (lastPage) => {
      if (!lastPage.nextCursorId || !lastPage.nextCursorCreatedAt) return
      return {
        cursorId: lastPage.nextCursorId,
        cursorCreatedAt: lastPage.nextCursorCreatedAt,
        query: query.trim(),
      }
    },
    enabled: category === "following" || !!query.trim(),
  })

  const {
    data: followerData,
    fetchNextPage: fetchFollowerNextPage,
    isFetchingNextPage: isFetchingFollowerNextPage,
    hasNextPage: hasFollowerNextPage,
    isPending: isFollowerPending,
  } = useInfiniteQuery({
    queryKey: getFollowServiceListMyFollowersQueryKey({ query: query.trim() }),
    initialPageParam: { query: query.trim() } as FollowServiceListMyFollowersParams,
    queryFn: ({ pageParam, signal }) => followServiceListMyFollowers(pageParam, undefined, signal),
    getNextPageParam: (lastPage) => {
      if (!lastPage.nextCursorId || !lastPage.nextCursorCreatedAt) return
      return {
        cursorId: lastPage.nextCursorId,
        cursorCreatedAt: lastPage.nextCursorCreatedAt,
        query: query.trim(),
      }
    },
    enabled: category === "followers" || !!query.trim(),
  })

  const followingUsers: User[] = followingData?.pages.flatMap((page) => page.users) ?? []
  const followerUsers: User[] = followerData?.pages.flatMap((page) => page.users) ?? []
  const isFollowingCategory = category === "following"
  const activeUsers = isFollowingCategory ? followingUsers : followerUsers
  const fetchNextPage = isFollowingCategory ? fetchFollowingNextPage : fetchFollowerNextPage
  const isFetchingNextPage = isFollowingCategory ? isFetchingFollowingNextPage : isFetchingFollowerNextPage
  const hasNextPage = isFollowingCategory ? hasFollowingNextPage : hasFollowerNextPage
  const isPending = isFollowingCategory ? isFollowingPending : isFollowerPending

  const listRef = useRef<HTMLDivElement>(null)
  // eslint-disable-next-line react-hooks/incompatible-library
  const virtualizer = useVirtualizer({
    count: activeUsers.length,
    estimateSize: () => 150,
    getScrollElement: () => listRef.current,
    getItemKey: (index) => activeUsers[index]?.uid ?? index,
    gap: 8,
    paddingStart: 4,
    paddingEnd: 4,
  })
  useEffect(() => {
    virtualizer.shouldAdjustScrollPositionOnItemSizeChange = (item, _delta, instance) => {
      const scrollOffset = instance.scrollOffset ?? 0
      return item.end <= scrollOffset
    }
  }, [virtualizer])
  const virtualItems = virtualizer.getVirtualItems()

  useEffect(() => {
    const lastVirtualItem = virtualItems.at(-1)
    if (!lastVirtualItem) return
    if (!hasNextPage || isFetchingNextPage) return
    if (lastVirtualItem.index < activeUsers.length - 1) return
    fetchNextPage()
  }, [activeUsers.length, fetchNextPage, hasNextPage, isFetchingNextPage, virtualItems])

  useEffect(() => {
    listRef.current?.scrollTo({ top: 0 })
  }, [category, query])

  const handleCategoryChange = (nextCategory: RelationCategory) => {
    const nextSearchParams = new URLSearchParams(searchParams)
    nextSearchParams.set("tab", nextCategory)
    setSearchParams(nextSearchParams)
  }

  const handleQueryChange = (nextQuery: string) => {
    const nextSearchParams = new URLSearchParams(searchParams)
    if (nextQuery.trim()) {
      nextSearchParams.set("query", nextQuery)
    } else {
      nextSearchParams.delete("query")
    }
    setSearchParams(nextSearchParams)
  }

  return (
    <div className="flex h-full min-h-0 justify-center gap-4 p-4">
      <RelationCategorySidenav selectedCategory={category} onCategoryChange={handleCategoryChange} className="h-full w-56" />
      <div className="flex min-h-0 w-full max-w-4xl flex-col gap-4">
        <RelationSearchCard query={query.trim()} onQueryChange={handleQueryChange} />
        {!activeUsers.length && !isPending ? (
          <div className="min-h-0 flex-1">
            <Empty className="h-full border">
              <EmptyHeader>
                <EmptyTitle>No Users</EmptyTitle>
                <EmptyDescription>No matching users were found.</EmptyDescription>
              </EmptyHeader>
            </Empty>
          </div>
        ) : (
          <div ref={listRef} className="min-h-0 flex-1 overflow-y-auto">
            {!activeUsers.length && isPending && <RelationListSkeleton />}
            {!!activeUsers.length && (
              <div className="relative" style={{ height: `${virtualizer.getTotalSize()}px` }}>
                {virtualItems.map((virtualItem) => (
                  <div
                    key={virtualItem.key}
                    ref={virtualizer.measureElement}
                    data-index={virtualItem.index}
                    className="absolute top-0 left-0 w-full"
                    style={{ transform: `translateY(${virtualItem.start}px)` }}
                  >
                    <RelationUserCard user={activeUsers[virtualItem.index]} relation={category} />
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
