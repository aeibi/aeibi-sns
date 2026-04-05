import { useUserServiceGetMe } from "@/api/generated"
import { PostListSkeleton } from "@/components/loading-skeleton"
import { PostCard } from "@/components/post-card"
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/components/ui/empty"
import { VirtualList } from "@/components/virtual-list"
import { useFavoritePostsFeed } from "@/hooks/use-post-infinite-feed"

export function Favorites() {
  const { data: userData } = useUserServiceGetMe()
  const { posts, fetchNextPage, isFetchingNextPage, hasNextPage, isPending, updatePostLocal, removePostLocal } = useFavoritePostsFeed()

  if (isPending && !posts.length) return <FavoritesSkeleton />
  if (!posts.length) return <FavoritesEmpty />
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

function FavoritesSkeleton() {
  return (
    <div className="h-full w-full overflow-y-auto">
      <PostListSkeleton count={3} />
    </div>
  )
}

function FavoritesEmpty() {
  return (
    <div className="h-full w-full p-4">
      <div className="mx-auto h-full w-full max-w-4xl px-4 py-4">
        <Empty className="h-full border">
          <EmptyHeader>
            <EmptyTitle>No Favorites Yet</EmptyTitle>
            <EmptyDescription>Posts you collect will appear here.</EmptyDescription>
          </EmptyHeader>
        </Empty>
      </div>
    </div>
  )
}
