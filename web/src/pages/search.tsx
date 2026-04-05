import { SearchForm } from "@/components/search-form"
import { useSearchParams } from "react-router-dom"
import { useUserServiceGetMe } from "@/api/generated"
import { SearchPageSkeleton } from "@/components/loading-skeleton"
import { PostCard } from "@/components/post-card"
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/components/ui/empty"
import { VirtualList } from "@/components/virtual-list"
import { useSearchPostsFeed } from "@/hooks/use-post-infinite-feed"

export function Search() {
  const { data: userData } = useUserServiceGetMe()

  const [searchParams] = useSearchParams()
  const query = searchParams.get("query") ?? ""
  const { posts, fetchNextPage, isFetchingNextPage, hasNextPage, isPending, updatePostLocal, removePostLocal } = useSearchPostsFeed(query)

  if (isPending && !posts.length) return <SearchSkeleton />
  if (!posts.length) return <SearchEmpty query={query} />
  return (
    <div className="h-full w-full">
      <VirtualList
        header={<SearchForm className="w-full" searchText={query} />}
        items={posts}
        getItemKey={(post) => post.uid}
        hasNextPage={hasNextPage}
        isFetchingNextPage={isFetchingNextPage}
        onLoadMore={fetchNextPage}
        renderItem={(post) => (
          <PostCard
            user={userData?.user}
            post={post}
            onUpdatePost={(patch) => updatePostLocal(post.uid, patch)}
            onRemovePost={() => removePostLocal(post.uid)}
          />
        )}
      />
    </div>
  )
}

function SearchSkeleton() {
  return (
    <div className="h-full w-full overflow-y-auto">
      <SearchPageSkeleton count={3} />
    </div>
  )
}

function SearchEmpty({ query }: { query: string }) {
  return (
    <div className="h-full w-full p-4">
      <div className="mx-auto flex h-full w-full max-w-4xl flex-col gap-4 px-4 py-4">
        <SearchForm className="w-full" searchText={query} />
        <Empty className="h-full border">
          <EmptyHeader>
            <EmptyTitle>{query.trim() ? "No Results" : "Start Searching"}</EmptyTitle>
            <EmptyDescription>
              {query.trim() ? "Try another keyword or check spelling." : "Enter keywords to search posts, users, or tags."}
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      </div>
    </div>
  )
}
