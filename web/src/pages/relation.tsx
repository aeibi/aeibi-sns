import { useInfiniteQuery } from "@tanstack/react-query"
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
import { VirtualList } from "@/components/virtual-list"
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
          <>
            {!activeUsers.length && isPending && (
              <div className="min-h-0 flex-1 overflow-y-auto">
                <RelationListSkeleton />
              </div>
            )}
            {!!activeUsers.length && (
              <VirtualList
                key={`${category}-${query.trim()}`}
                items={activeUsers}
                getItemKey={(user) => user.uid}
                hasNextPage={hasNextPage}
                isFetchingNextPage={isFetchingNextPage}
                onLoadMore={fetchNextPage}
                estimateSize={() => 150}
                gap={8}
                paddingStart={4}
                paddingEnd={4}
                className="min-h-0 flex-1 overflow-y-auto"
                innerClassName="w-full"
                renderItem={(user) => <RelationUserCard user={user} relation={category} />}
              />
            )}
          </>
        )}
      </div>
    </div>
  )
}
