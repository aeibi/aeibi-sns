import { SearchForm } from "@/components/search-form"
import { useSearchParams } from "react-router-dom"
import { useUserServiceGetMe } from "@/api/generated"
import { PostCard } from "@/components/post-card"
import { VirtualList } from "@/components/virtual-list"
import { useSearchPostsFeed } from "@/hooks/use-post-infinite-feed"

export function Search() {
  const { data: userData } = useUserServiceGetMe()
  const [searchParams] = useSearchParams()
  const query = searchParams.get("query") ?? ""
  const { posts, fetchNextPage, isFetchingNextPage, hasNextPage, updatePostLocal, removePostLocal } = useSearchPostsFeed({
    query,
  })
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
