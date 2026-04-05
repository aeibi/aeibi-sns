import { SearchForm } from "@/components/search-form"
import { useSearchParams } from "react-router-dom"
import { useUserServiceGetMe } from "@/api/generated"
import { PostCard } from "@/components/post-card"
import { VirtualList } from "@/components/virtual-list"
import { usePostsFeed } from "@/hooks/use-post-infinite-feed"

export function Tag() {
  const { data: userData } = useUserServiceGetMe()
  const [searchParams] = useSearchParams()
  const tagName = searchParams.get("tag") ?? ""
  const { posts, fetchNextPage, isFetchingNextPage, hasNextPage, updatePostLocal, removePostLocal } = usePostsFeed({
    tagName,
  })
  return (
    <div className="h-full w-full">
      <VirtualList
        header={<SearchForm className="w-full" searchText={tagName} />}
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
