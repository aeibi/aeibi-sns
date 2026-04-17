import { useInfiniteQuery, useQueryClient, type InfiniteData, type QueryKey } from "@tanstack/react-query"
import { dedupeByUid } from "@/lib/utils"
import type { Comment } from "@/types/comment"
import {
  commentServiceListTopComments,
  getCommentServiceListTopCommentsQueryKey,
  type CommentListTopCommentsResponse,
  type CommentServiceListTopCommentsParams,
} from "@/api/generated"

interface UseCommentInfiniteFeedOptions {
  queryKey: QueryKey
  initialPageParam: CommentServiceListTopCommentsParams
  enabled?: boolean
  queryFn: (pageParam: CommentServiceListTopCommentsParams, signal?: AbortSignal) => Promise<CommentListTopCommentsResponse>
  getNextPageParam: (lastPage: CommentListTopCommentsResponse) => CommentServiceListTopCommentsParams | undefined
}

export function useCommentInfiniteFeed({
  queryKey,
  initialPageParam,
  enabled = true,
  queryFn,
  getNextPageParam,
}: UseCommentInfiniteFeedOptions) {
  const queryClient = useQueryClient()
  const query = useInfiniteQuery<
    CommentListTopCommentsResponse,
    unknown,
    InfiniteData<CommentListTopCommentsResponse>,
    QueryKey,
    CommentServiceListTopCommentsParams
  >({
    queryKey,
    enabled,
    initialPageParam,
    queryFn: ({ pageParam, signal }) => queryFn(pageParam, signal),
    getNextPageParam,
    gcTime: 0,
  })

  function addCommentLocal(comment: Comment) {
    queryClient.setQueryData<InfiniteData<CommentListTopCommentsResponse>>(queryKey, (old) => {
      if (!old || !old.pages.length) return old
      const firstPage = old.pages[0]
      if (firstPage.comments.some((item) => item.uid === comment.uid)) return old
      const pages = [{ ...firstPage, comments: [comment, ...firstPage.comments] }, ...old.pages.slice(1)]
      return { ...old, pages }
    })
  }

  function updateCommentLocal(uid: string, patch: Partial<Comment>) {
    queryClient.setQueryData<InfiniteData<CommentListTopCommentsResponse>>(queryKey, (old) => {
      if (!old) return old
      let changed = false
      const pages = old.pages.map((page) => {
        let pageChanged = false
        const comments = page.comments.map((comment) => {
          if (comment.uid !== uid) return comment
          pageChanged = true
          changed = true
          return { ...comment, ...patch }
        })
        return pageChanged ? { ...page, comments } : page
      })
      return changed ? { ...old, pages } : old
    })
  }

  function removeCommentLocal(uid: string) {
    queryClient.setQueryData<InfiniteData<CommentListTopCommentsResponse>>(queryKey, (old) => {
      if (!old) return old
      let changed = false
      const pages = old.pages.map((page) => {
        const comments = page.comments.filter((comment) => comment.uid !== uid)
        if (comments.length !== page.comments.length) {
          changed = true
          return { ...page, comments }
        }
        return page
      })
      return changed ? { ...old, pages } : old
    })
  }

  const comments = dedupeByUid(query.data?.pages.flatMap((page) => page.comments) ?? [])

  return { ...query, comments, queryKey, addCommentLocal, updateCommentLocal, removeCommentLocal }
}

export function useTopCommentsFeed(postUid: string) {
  const queryKey = [...getCommentServiceListTopCommentsQueryKey(postUid), "infinite"] as const
  return useCommentInfiniteFeed({
    queryKey,
    enabled: !!postUid,
    initialPageParam: {},
    queryFn: (pageParam, signal) => commentServiceListTopComments(postUid, pageParam, undefined, signal),
    getNextPageParam: (lastPage) => {
      if (!lastPage.nextPageToken) return
      return {
        pageToken: lastPage.nextPageToken,
      }
    },
  })
}
