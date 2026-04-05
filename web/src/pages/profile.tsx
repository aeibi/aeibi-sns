import { useSearchParams } from "react-router-dom"
import { useFollowServiceFollow, useUserServiceGetMe, useUserServiceGetUser } from "@/api/generated"
import { useQueryClient } from "@tanstack/react-query"
import { PostCard } from "@/components/post-card"
import { ProfileCard } from "@/components/profile-card"
import { VirtualList } from "@/components/virtual-list"
import { toast } from "sonner"
import { usePostsFeed } from "@/hooks/use-post-infinite-feed"
import { EmptyState } from "@/components/empty"

export function Profile() {
  const queryClient = useQueryClient()

  const [searchParams] = useSearchParams()
  const { data: meData } = useUserServiceGetMe()
  const uid = searchParams.get("uid") || meData?.user.uid || ""
  const { data: userData, queryKey, isPending: isUserPending } = useUserServiceGetUser(uid)
  const { mutate: followUser, isPending: isFollowPending } = useFollowServiceFollow()
  const { posts, fetchNextPage, isFetchingNextPage, hasNextPage, updatePostLocal, removePostLocal } = usePostsFeed({
    authorUid: uid,
  })
  const handleFollow = () => {
    const user = userData?.user
    if (!user || !meData?.user || meData.user.uid === user.uid || isFollowPending) return
    followUser(
      {
        uid: user.uid,
        data: { uid: user.uid, action: Number(user.isFollowing) },
      },
      {
        onSuccess: () => {
          queryClient.setQueryData(queryKey, (oldData) => {
            if (!oldData) return oldData
            return {
              ...oldData,
              user: {
                ...oldData.user,
                isFollowing: !user.isFollowing,
                followersCount: Math.max(0, user.followersCount + (user.isFollowing ? -1 : 1)),
              },
            }
          })
        },
        onError: () => toast.error("Failed to update follow status.", { position: "top-center" }),
      }
    )
  }
  if (isUserPending) return null
  if (!userData) return <EmptyState />
  return (
    <div className="h-full w-full">
      <VirtualList
        header={
          <ProfileCard className="w-full" user={userData?.user} me={meData?.user} onFollow={handleFollow} followPending={isFollowPending} />
        }
        items={posts}
        getItemKey={(post) => post.uid}
        hasNextPage={hasNextPage}
        isFetchingNextPage={isFetchingNextPage}
        onLoadMore={fetchNextPage}
        renderItem={(post) => (
          <PostCard
            user={meData?.user}
            post={post}
            onUpdatePost={(patch) => updatePostLocal(post.uid, patch)}
            onRemovePost={() => removePostLocal(post.uid)}
          />
        )}
      />
    </div>
  )
}
