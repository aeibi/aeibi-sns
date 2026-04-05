import { type ReactNode } from "react"
import { commentServiceGetComment, useUserServiceGetMe } from "@/api/generated"
import { PostCommentsComposer } from "@/components/post-comment-composer"
import { PostComment } from "@/components/post-comment"
import { PostCard } from "@/components/post-card"
import { PostCommentsPreviewSkeleton, PostDetailSkeleton } from "@/components/loading-skeleton"
import { VirtualList } from "@/components/virtual-list"
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/components/ui/empty"
import { Separator } from "@/components/ui/separator"
import { useTopCommentsFeed } from "@/hooks/use-comment-infinite-feed"
import { usePostFeed } from "@/hooks/use-post-feed"
import { ChevronLeftIcon } from "lucide-react"
import { useNavigate, useOutlet, useParams } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { toast } from "sonner"

type PostDetailHostRouteProps = {
  children: ReactNode
}

export function PostDetailHostRoute({ children }: PostDetailHostRouteProps) {
  const outlet = useOutlet()

  return (
    <div className="relative h-full min-h-0">
      {children}
      {!!outlet && <div className="absolute inset-0 z-20 min-h-0">{outlet}</div>}
    </div>
  )
}

export function PostDetail() {
  const navigate = useNavigate()
  const { id = "" } = useParams()

  const { data: userData } = useUserServiceGetMe()
  const { post, isPending: isPostPending, updatePostLocal } = usePostFeed(id)
  const {
    comments,
    fetchNextPage,
    isFetchingNextPage,
    hasNextPage,
    isPending: isCommentsPending,
    addCommentLocal,
    updateCommentLocal,
    removeCommentLocal,
  } = useTopCommentsFeed(id)

  const handlePosted = (uid: string) => {
    if (post) updatePostLocal(post.uid, { commentCount: post.commentCount + 1 })
    void commentServiceGetComment(uid)
      .then((data) => {
        addCommentLocal(data.comment)
      })
      .catch(() => toast.error("Failed to load created comment.", { position: "top-center" }))
  }

  if (isPostPending) {
    return (
      <div className="flex h-full min-h-full w-full">
        <div className="h-full w-full overflow-y-auto bg-background py-16">
          <div className="mx-auto flex w-full max-w-4xl items-center px-4 pb-4">
            <Button variant="outline" size="icon" onClick={() => navigate(-1)}>
              <ChevronLeftIcon />
            </Button>
          </div>
          <PostDetailSkeleton />
        </div>
      </div>
    )
  }

  if (!post) {
    return (
      <div className="flex h-full min-h-full w-full">
        <div className="h-full w-full bg-background p-4">
          <div className="mx-auto flex w-full max-w-4xl items-center pb-4">
            <Button variant="outline" size="icon" onClick={() => navigate(-1)}>
              <ChevronLeftIcon />
            </Button>
          </div>
          <div className="mx-auto h-[calc(100%-3.5rem)] w-full max-w-4xl">
            <Empty className="h-full border">
              <EmptyHeader>
                <EmptyTitle>Post Not Found</EmptyTitle>
                <EmptyDescription>The post may have been deleted or the link is invalid.</EmptyDescription>
              </EmptyHeader>
            </Empty>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="flex h-full min-h-full w-full">
      <div className="h-full w-full bg-background">
        <VirtualList
          stickyHeader={
            <div className="sticky top-0 z-10 mx-auto flex max-w-4xl items-center p-4">
              <Button variant="outline" size="icon" onClick={() => navigate(-1)}>
                <ChevronLeftIcon />
              </Button>
            </div>
          }
          header={
            <div className="flex flex-col gap-4">
              <PostCard
                post={post}
                user={userData?.user}
                onUpdatePost={(patch) => updatePostLocal(post.uid, patch)}
                onRemovePost={() => navigate("/")}
                disableCommentExpand
              />
              {!!userData?.user && (
                <PostCommentsComposer className="w-full" user={userData.user} postUid={post.uid} onPosted={handlePosted} />
              )}
              {isCommentsPending && !comments.length && <PostCommentsPreviewSkeleton count={3} />}
            </div>
          }
          items={comments}
          getItemKey={(comment) => comment.uid}
          hasNextPage={hasNextPage}
          isFetchingNextPage={isFetchingNextPage}
          onLoadMore={fetchNextPage}
          paddingStart={64}
          renderItem={(comment) => (
            <div className="flex flex-col gap-4">
              <PostComment
                comment={comment}
                user={userData?.user}
                onUpdateComment={(patch) => updateCommentLocal(comment.uid, patch)}
                onRemoveComment={() => removeCommentLocal(comment.uid)}
              />
              <Separator className="ml-8" />
            </div>
          )}
        />
      </div>
    </div>
  )
}
