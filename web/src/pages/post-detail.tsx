import { type ReactNode } from "react"
import { commentServiceGetComment, useUserServiceGetMe } from "@/api/generated"
import { PostCommentsComposer } from "@/components/post-comment-composer"
import { PostComment } from "@/components/post-comment"
import { PostCard } from "@/components/post-card"
import { VirtualList } from "@/components/virtual-list"
import { Separator } from "@/components/ui/separator"
import { useTopCommentsFeed } from "@/hooks/use-comment-infinite-feed"
import { usePostFeed } from "@/hooks/use-post-feed"
import { ChevronLeftIcon } from "lucide-react"
import { useNavigate, useOutlet, useParams } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { toast } from "sonner"
import { EmptyState } from "@/components/empty"

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
  const { post, updatePostLocal, isPending: isPostPending } = usePostFeed(id)
  const { comments, fetchNextPage, isFetchingNextPage, hasNextPage, addCommentLocal, updateCommentLocal, removeCommentLocal } =
    useTopCommentsFeed(id)

  const handlePosted = (uid: string) => {
    if (post) updatePostLocal(post.uid, { commentCount: post.commentCount + 1 })
    void commentServiceGetComment(uid)
      .then((data) => {
        addCommentLocal(data.comment)
      })
      .catch(() => toast.error("Failed to load created comment.", { position: "top-center" }))
  }
  if (isPostPending) return null
  if (!post) return <EmptyState />
  return (
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
  )
}
