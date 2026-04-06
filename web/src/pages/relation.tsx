import { useInfiniteQuery } from "@tanstack/react-query"
import { useSearchParams } from "react-router-dom"
import {
  followServiceListMyFollowers,
  followServiceListMyFollowing,
  getFollowServiceListMyFollowersQueryKey,
  getFollowServiceListMyFollowingQueryKey,
} from "@/api/generated"
import { RelationCategorySidenav, type RelationCategory } from "@/components/relation-category-sidenav"
import { RelationSearchCard } from "@/components/relation-search-card"
import { VirtualList } from "@/components/virtual-list"
import { RelationUserCard } from "@/components/relation-user-card"
import type { User } from "@/types/user"
import { dedupeByUid } from "@/lib/utils"

export function Relation() {
  const [searchParams, setSearchParams] = useSearchParams()
  const category = searchParams.get("tab") === "followers" ? "followers" : "following"
  const query = searchParams.get("query") ?? ""
  const normalizedQuery = query.trim()

  const {
    data: relationData,
    fetchNextPage,
    isFetchingNextPage,
    hasNextPage,
  } = useInfiniteQuery({
    queryKey:
      category === "following"
        ? getFollowServiceListMyFollowingQueryKey({ query: normalizedQuery })
        : getFollowServiceListMyFollowersQueryKey({ query: normalizedQuery }),
    initialPageParam: { query: normalizedQuery },
    queryFn: ({ pageParam, signal }) =>
      category === "following"
        ? followServiceListMyFollowing(pageParam, undefined, signal)
        : followServiceListMyFollowers(pageParam, undefined, signal),
    getNextPageParam: (lastPage) => {
      if (!lastPage.nextPageToken) return
      return {
        pageToken: lastPage.nextPageToken,
        query: normalizedQuery,
      }
    },
  })

  const users: User[] = dedupeByUid(relationData?.pages.flatMap((page) => page.users) ?? [])

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
        <RelationSearchCard query={normalizedQuery} onQueryChange={handleQueryChange} />
        <VirtualList
          key={`${category}-${normalizedQuery}`}
          items={users}
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
      </div>
    </div>
  )
}
