import { postServiceGetPost, useUserServiceGetMe } from "@/api/generated"
import { PostCard } from "@/components/post-card"
import { PostComposerCard } from "@/components/post-composer"
import { VirtualList } from "@/components/virtual-list"
import { useHomePostsFeed } from "@/hooks/use-post-infinite-feed"
import { toast } from "sonner"

export function Home() {
  const { data: userData } = useUserServiceGetMe()
  const { posts, fetchNextPage, isFetchingNextPage, hasNextPage, addPostLocal, updatePostLocal, removePostLocal } = useHomePostsFeed()
  const handlePosted = (uid: string) => {
    void postServiceGetPost(uid)
      .then((data) => addPostLocal(data.post))
      .catch(() => toast.error("Failed to load created post.", { position: "top-center" }))
  }
  return (
    <div className="h-full w-full">
      <VirtualList
        header={!!userData?.user && <PostComposerCard onPosted={handlePosted} className="w-full" />}
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
