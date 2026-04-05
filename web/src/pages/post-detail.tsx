import { useEffect, useRef, useState, type ReactNode } from "react"
import { defaultRangeExtractor, useVirtualizer } from "@tanstack/react-virtual"
import {
  commentServiceGetComment,
  commentServiceListTopComments,
  type CommentListTopCommentsResponse,
  type CommentServiceListTopCommentsParams,
  getCommentServiceListTopCommentsQueryKey,
  getPostServiceGetPostQueryKey,
  getUserServiceGetUserQueryKey,
  type PostGetPostResponse,
  useFollowServiceFollow,
  usePostServiceCollectPost,
  usePostServiceDeletePost,
  usePostServiceGetPost,
  usePostServiceLikePost,
  useUserServiceGetMe,
} from "@/api/generated"
import { useInfiniteQuery, useQueryClient, type InfiniteData } from "@tanstack/react-query"
import { PostCommentsComposer } from "@/components/post-comment-composer"
import { PostComment } from "@/components/post-comment"
import { PostActionButton } from "@/components/post-action"
import { PostCommentsPreviewSkeleton, PostDetailSkeleton } from "@/components/loading-skeleton"
import { PostHeader } from "@/components/post-header"
import { PostMedia } from "@/components/post-media"
import { PostTag } from "@/components/post-tag"
import { PostText } from "@/components/post-text"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/components/ui/empty"
import { Separator } from "@/components/ui/separator"
import { useCommentListLocalActions } from "@/hooks/use-comment-list-local-actions"
import { dedupeByUid } from "@/lib/utils"
import { BookmarkIcon, ChevronLeftIcon, HeartIcon, MessageCircleIcon } from "lucide-react"
import { useNavigate, useOutlet, useParams } from "react-router-dom"
import { Button } from "@/components/ui/button"
import type { Post } from "@/types/post"
import type { User } from "@/types/user"
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
  const queryClient = useQueryClient()

  const { data: userData } = useUserServiceGetMe()
  const { data: postData, isPending: isPostPending } = usePostServiceGetPost(id)
  const post = postData?.post

  const queryKey = [...getCommentServiceListTopCommentsQueryKey(id), "infinite"] as const
  const {
    data,
    fetchNextPage,
    isFetchingNextPage,
    hasNextPage,
    isPending: isCommentsPending,
  } = useInfiniteQuery({
    queryKey,
    enabled: !!id,
    initialPageParam: {} as CommentServiceListTopCommentsParams,
    queryFn: ({ pageParam, signal }) => commentServiceListTopComments(id, pageParam, undefined, signal),
    getNextPageParam: (lastPage) => {
      if (!lastPage.nextCursorId || !lastPage.nextCursorCreatedAt) return
      return { cursorId: lastPage.nextCursorId, cursorCreatedAt: lastPage.nextCursorCreatedAt }
    },
  })
  const comments = dedupeByUid(data?.pages.flatMap((p) => p.comments) ?? [])
  const { updateCommentLocal, removeCommentLocal } = useCommentListLocalActions<CommentListTopCommentsResponse>(queryKey)

  const handlePosted = (uid: string) => {
    queryClient.setQueryData<PostGetPostResponse>(getPostServiceGetPostQueryKey(id), (old) => {
      if (!old) return old
      return {
        ...old,
        post: {
          ...old.post,
          commentCount: old.post.commentCount + 1,
        },
      }
    })
    void commentServiceGetComment(uid)
      .then((data) => {
        queryClient.setQueryData<InfiniteData<CommentListTopCommentsResponse>>(queryKey, (old) => {
          if (!old || !old.pages.length) return old
          const firstPage = old.pages[0]
          if (firstPage.comments.some((comment) => comment.uid === data.comment.uid)) return old
          const pages = [{ ...firstPage, comments: [data.comment, ...firstPage.comments] }, ...old.pages.slice(1)]
          return { ...old, pages }
        })
      })
      .catch(() => toast.error("Failed to load created comment.", { position: "top-center" }))
  }

  const ref = useRef<HTMLDivElement>(null)

  // eslint-disable-next-line react-hooks/incompatible-library
  const virtualizer = useVirtualizer({
    count: comments.length + 1,
    estimateSize: () => 300,
    getScrollElement: () => ref.current,
    getItemKey: (index) => (index === 0 ? "post" : (comments[index - 1]?.uid ?? index)),
    rangeExtractor: (range) => {
      const next = new Set([0, ...defaultRangeExtractor(range)])
      return [...next].sort((a, b) => a - b)
    },
    gap: 16,
    paddingStart: 64,
    paddingEnd: 16,
  })
  useEffect(() => {
    virtualizer.shouldAdjustScrollPositionOnItemSizeChange = (item, _delta, instance) => {
      const scrollOffset = instance.scrollOffset ?? 0
      return item.end <= scrollOffset
    }
  }, [virtualizer])
  const virtualItems = virtualizer.getVirtualItems()

  useEffect(() => {
    const lastCommentItem = [...virtualItems].reverse().find((item) => item.index > 0)
    if (!lastCommentItem) return
    if (!hasNextPage || isFetchingNextPage) return
    if (lastCommentItem.index < comments.length) return
    fetchNextPage()
  }, [comments.length, fetchNextPage, hasNextPage, isFetchingNextPage, virtualItems])

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
        <div ref={ref} className="h-full w-full overflow-y-auto">
          <div className="relative" style={{ height: `${virtualizer.getTotalSize()}px` }}>
            <div className="sticky top-0 z-10 mx-auto flex max-w-4xl items-center p-4">
              <Button variant="outline" size="icon" onClick={() => navigate(-1)}>
                <ChevronLeftIcon />
              </Button>
            </div>
            {virtualItems.map((virtualItem) => {
              const comment = virtualItem.index > 0 ? comments[virtualItem.index - 1] : undefined
              return (
                <div
                  key={virtualItem.key}
                  ref={virtualizer.measureElement}
                  data-index={virtualItem.index}
                  className="absolute top-0 left-0 w-full"
                  style={{ transform: `translateY(${virtualItem.start}px)` }}
                >
                  <div className="mx-auto w-full max-w-4xl px-4">
                    {virtualItem.index === 0 ? (
                      <div className="flex flex-col gap-4">
                        {!!post && <PostDetailCard key={post.uid} post={post} user={userData?.user} />}
                        {!!userData?.user && !!post && (
                          <PostCommentsComposer className="w-full" user={userData.user} postUid={post.uid} onPosted={handlePosted} />
                        )}
                        {isCommentsPending && !comments.length && <PostCommentsPreviewSkeleton count={3} />}
                      </div>
                    ) : (
                      !!comment && (
                        <div className="flex flex-col gap-4">
                          <PostComment
                            key={comment.uid}
                            comment={comment}
                            user={userData?.user}
                            onUpdateComment={(patch) => updateCommentLocal(comment.uid, patch)}
                            onRemoveComment={() => removeCommentLocal(comment.uid)}
                          />
                          <Separator className="ml-8" />
                        </div>
                      )
                    )}
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      </div>
    </div>
  )
}

function PostDetailCard({ post, user }: { post: Post; user?: User }) {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const isOwnPost = !!user && user.uid === post.author.uid
  const [author, setAuthor] = useState(post.author)
  const [interactions, setInteractions] = useState({
    liked: post.liked,
    likeCount: post.likeCount,
    collected: post.collected,
    collectionCount: post.collectionCount,
  })

  const { mutate: deletePost, isPending: isDeleting } = usePostServiceDeletePost()
  const { mutate: likePost, isPending: isLiking } = usePostServiceLikePost()
  const { mutate: collectPost, isPending: isCollecting } = usePostServiceCollectPost()
  const { mutate: followUser, isPending: isFollowPending } = useFollowServiceFollow()

  const handleDelete = () => {
    if (isDeleting) return
    deletePost(
      { uid: post.uid },
      {
        onSuccess: () => navigate("/"),
        onError: () => toast.error("Failed to delete the post.", { position: "top-center" }),
      }
    )
  }

  const handleLike = () => {
    if (!user || isLiking) return
    const previous = interactions
    const next = {
      ...interactions,
      liked: !interactions.liked,
      likeCount: Math.max(0, interactions.likeCount + (interactions.liked ? -1 : 1)),
    }
    setInteractions(next)
    likePost(
      {
        uid: post.uid,
        data: { uid: post.uid, action: Number(interactions.liked) },
      },
      {
        onError: () => {
          toast.error("Failed to like the post.", { position: "top-center" })
          setInteractions(previous)
        },
      }
    )
  }

  const handleCollect = () => {
    if (!user || isCollecting) return
    const previous = interactions
    const next = {
      ...interactions,
      collected: !interactions.collected,
      collectionCount: Math.max(0, interactions.collectionCount + (interactions.collected ? -1 : 1)),
    }
    setInteractions(next)
    collectPost(
      {
        uid: post.uid,
        data: { uid: post.uid, action: Number(interactions.collected) },
      },
      {
        onError: () => {
          toast.error("Failed to collect the post.", { position: "top-center" })
          setInteractions(previous)
        },
      }
    )
  }

  const handleFollow = () => {
    if (!user || isOwnPost || isFollowPending) return
    const previousAuthor = author
    setAuthor({
      ...previousAuthor,
      isFollowing: !previousAuthor.isFollowing,
    })
    followUser(
      {
        uid: previousAuthor.uid,
        data: { uid: previousAuthor.uid, action: Number(previousAuthor.isFollowing) },
      },
      {
        onSuccess: () => {
          void queryClient.invalidateQueries({ queryKey: getUserServiceGetUserQueryKey(previousAuthor.uid) })
        },
        onError: () => {
          setAuthor(previousAuthor)
          toast.error("Failed to update follow status.", { position: "top-center" })
        },
      }
    )
  }

  return (
    <Card>
      <CardHeader>
        <PostHeader post={{ ...post, author }} isOwnPost={isOwnPost} onDelete={handleDelete} onFollow={handleFollow} />
      </CardHeader>
      <CardContent className="flex flex-col gap-3">
        <PostText post={post} />
        <PostMedia post={post} />
        <PostTag tags={post.tags} />
        <div className="flex items-center gap-1 text-muted-foreground">
          <PostActionButton icon={<MessageCircleIcon />} count={post.commentCount} />
          <PostActionButton
            icon={<HeartIcon />}
            count={interactions.likeCount}
            active={interactions.liked}
            disabled={!user || isLiking}
            onClick={handleLike}
          />
          <PostActionButton
            icon={<BookmarkIcon />}
            count={interactions.collectionCount}
            active={interactions.collected}
            disabled={!user || isCollecting}
            onClick={handleCollect}
          />
        </div>
      </CardContent>
    </Card>
  )
}
