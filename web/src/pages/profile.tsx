import { useSearchParams } from "react-router-dom"
import {
  getUserServiceGetUserQueryKey,
  useFollowServiceFollow,
  useUserServiceGetMe,
  useUserServiceGetUser,
  type UserGetUserResponse,
} from "@/api/generated"
import { useQueryClient } from "@tanstack/react-query"
import { ProfilePageSkeleton } from "@/components/loading-skeleton"
import { PostFeedList } from "@/components/post-feed-list"
import { ProfileCard } from "@/components/profile-card"
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/components/ui/empty"
import { toast } from "sonner"
import type { User } from "@/types/user"
import { useAuthorPostsFeed } from "@/hooks/use-post-infinite-feed"

export function Profile() {
  const queryClient = useQueryClient()

  const [searchParams] = useSearchParams()
  const { data: meData } = useUserServiceGetMe()
  const uid = searchParams.get("uid") || meData?.user.uid || ""
  const { data: userData, isPending: isUserPending } = useUserServiceGetUser(uid)
  const { mutate: followUser, isPending: isFollowPending } = useFollowServiceFollow()
  const {
    posts,
    fetchNextPage,
    isFetchingNextPage,
    hasNextPage,
    isPending: isPostsPending,
    updatePostLocal,
    removePostLocal,
  } = useAuthorPostsFeed(uid)

  const handleFollow = () => {
    const user = userData?.user
    if (!user || !meData?.user || meData.user.uid === user.uid || isFollowPending) return
    const nextUser: User = {
      ...user,
      isFollowing: !user.isFollowing,
      followersCount: Math.max(0, user.followersCount + (user.isFollowing ? -1 : 1)),
    }
    followUser(
      {
        uid: user.uid,
        data: { uid: user.uid, action: Number(user.isFollowing) },
      },
      {
        onSuccess: () => {
          queryClient.setQueryData<UserGetUserResponse | undefined>(getUserServiceGetUserQueryKey(user.uid), (oldData) => {
            if (!oldData) return oldData
            return {
              ...oldData,
              user: nextUser,
            }
          })
        },
        onError: () => toast.error("Failed to update follow status.", { position: "top-center" }),
      }
    )
  }

  if (!!uid && (isUserPending || isPostsPending)) return <ProfileSkeleton />

  if (!uid && !meData?.user.uid) return <ProfileEmpty isLogged={false} />
  if (!userData) return <ProfileEmpty isLogged={!!uid} />
  return (
    <PostFeedList
      posts={posts}
      user={meData?.user}
      headerKey="profile"
      header={
        <ProfileCard className="w-full" user={userData.user} me={meData?.user} onFollow={handleFollow} followPending={isFollowPending} />
      }
      hasNextPage={hasNextPage}
      isFetchingNextPage={isFetchingNextPage}
      onLoadMore={fetchNextPage}
      onRemovePost={removePostLocal}
      onUpdatePost={updatePostLocal}
    />
  )
}

function ProfileSkeleton() {
  return (
    <div className="h-full w-full overflow-y-auto">
      <ProfilePageSkeleton count={2} />
    </div>
  )
}

function ProfileEmpty(isLogged: { isLogged: boolean }) {
  return (
    <div className="h-full w-full p-4">
      <div className="mx-auto h-full w-full max-w-4xl px-4 py-4">
        <Empty className="h-full border">
          <EmptyHeader>
            <EmptyTitle>{isLogged ? "User Not Found" : "Profile Unavailable"}</EmptyTitle>
            <EmptyDescription>
              {isLogged ? "The requested user does not exist or is unavailable." : "Please log in to view your profile."}
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      </div>
    </div>
  )
}
