import { useUserServiceGetMe } from "@/api/generated"
import { PostCard } from "@/components/post-card"
import { VirtualList } from "@/components/virtual-list"
import { useFavoritePostsFeed } from "@/hooks/use-post-infinite-feed"

export function Favorites() {
  const { data: userData } = useUserServiceGetMe()
  const { posts, fetchNextPage, isFetchingNextPage, hasNextPage, updatePostLocal, removePostLocal } = useFavoritePostsFeed()
  return (
    <div className="h-full w-full">
      <VirtualList
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
