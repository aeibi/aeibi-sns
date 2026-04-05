import { postServiceGetPost, useUserServiceGetMe } from "@/api/generated"
import { PostListSkeleton } from "@/components/loading-skeleton"
import { PostComposerCard } from "@/components/post-composer"
import { PostFeedList } from "@/components/post-feed-list"
import { useHomePostsFeed } from "@/hooks/use-post-infinite-feed"
import { toast } from "sonner"

export function Home() {
  const { data: userData } = useUserServiceGetMe()
  const { posts, fetchNextPage, isFetchingNextPage, hasNextPage, isPending, addPostLocal, updatePostLocal, removePostLocal } =
    useHomePostsFeed()
  const handlePosted = (uid: string) => {
    void postServiceGetPost(uid)
      .then((data) => addPostLocal(data.post))
      .catch(() => toast.error("Failed to load created post.", { position: "top-center" }))
  }

  if (isPending && !posts.length) return <HomeSkeleton />
  return (
    <PostFeedList
      posts={posts}
      user={userData?.user}
      headerKey="composer"
      header={!!userData?.user && <PostComposerCard onPosted={handlePosted} className="w-full" />}
      hasNextPage={hasNextPage}
      isFetchingNextPage={isFetchingNextPage}
      onLoadMore={fetchNextPage}
      onRemovePost={removePostLocal}
      onUpdatePost={updatePostLocal}
    />
  )
}

function HomeSkeleton() {
  return (
    <div className="h-full w-full overflow-y-auto">
      <PostListSkeleton count={3} />
    </div>
  )
}
