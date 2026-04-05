import { useInfiniteQuery, useQueryClient, type InfiniteData, type QueryKey } from "@tanstack/react-query"
import { dedupeByUid } from "@/lib/utils"
import type { Comment } from "@/types/comment"
import {
  commentServiceListTopComments,
  getCommentServiceListTopCommentsQueryKey,
  type CommentListTopCommentsResponse,
  type CommentServiceListTopCommentsParams,
} from "@/api/generated"

type CommentPage = { comments: Comment[] }
type CursorPage = { nextCursorId?: string; nextCursorCreatedAt?: string }
type CursorPageParam = { cursorId?: string; cursorCreatedAt?: string }

interface UseCommentInfiniteFeedOptions<TPage extends CommentPage, TPageParam, TQueryKey extends QueryKey> {
  queryKey: TQueryKey
  initialPageParam: TPageParam
  enabled?: boolean
  queryFn: (pageParam: TPageParam, signal?: AbortSignal) => Promise<TPage>
  getNextPageParam: (lastPage: TPage) => TPageParam | undefined
}

export function useCommentInfiniteFeed<TPage extends CommentPage, TPageParam, TQueryKey extends QueryKey = QueryKey>({
  queryKey,
  initialPageParam,
  enabled = true,
  queryFn,
  getNextPageParam,
}: UseCommentInfiniteFeedOptions<TPage, TPageParam, TQueryKey>) {
  const queryClient = useQueryClient()
  const query = useInfiniteQuery<TPage, unknown, InfiniteData<TPage>, TQueryKey, TPageParam>({
    queryKey,
    enabled,
    initialPageParam,
    queryFn: ({ pageParam, signal }) => queryFn(pageParam as TPageParam, signal),
    getNextPageParam,
  })

  function addCommentLocal(comment: Comment) {
    queryClient.setQueryData<InfiniteData<TPage>>(queryKey, (old) => {
      if (!old || !old.pages.length) return old
      const firstPage = old.pages[0]
      if (firstPage.comments.some((item) => item.uid === comment.uid)) return old
      const pages = [{ ...firstPage, comments: [comment, ...firstPage.comments] }, ...old.pages.slice(1)]
      return { ...old, pages }
    })
  }

  function updateCommentLocal(uid: string, patch: Partial<Comment>) {
    queryClient.setQueryData<InfiniteData<TPage>>(queryKey, (old) => {
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
    queryClient.setQueryData<InfiniteData<TPage>>(queryKey, (old) => {
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

export function buildCommentCursorNextPageParam<TPage extends CursorPage, TPageParam extends CursorPageParam>(
  lastPage: TPage,
  fixedParams?: Omit<TPageParam, "cursorId" | "cursorCreatedAt">,
  extraParams?: Partial<TPageParam>
): TPageParam | undefined {
  if (!lastPage.nextCursorId || !lastPage.nextCursorCreatedAt) return
  return {
    ...(fixedParams ?? ({} as Omit<TPageParam, "cursorId" | "cursorCreatedAt">)),
    ...(extraParams ?? {}),
    cursorId: lastPage.nextCursorId,
    cursorCreatedAt: lastPage.nextCursorCreatedAt,
  } as TPageParam
}

export function useTopCommentsFeed(postUid: string) {
  const queryKey = [...getCommentServiceListTopCommentsQueryKey(postUid), "infinite"] as const
  return useCommentInfiniteFeed<CommentListTopCommentsResponse, CommentServiceListTopCommentsParams>({
    queryKey,
    enabled: !!postUid,
    initialPageParam: {} as CommentServiceListTopCommentsParams,
    queryFn: (pageParam, signal) => commentServiceListTopComments(postUid, pageParam, undefined, signal),
    getNextPageParam: (lastPage) => buildCommentCursorNextPageParam(lastPage),
  })
}
