import { useInfiniteQuery, useQueryClient, type InfiniteData, type QueryKey } from "@tanstack/react-query"
import { dedupeByUid } from "@/lib/utils"
import type { Post } from "@/types/post"
import {
  getPostServiceListMyCollectionsQueryKey,
  getPostServiceListPostsByAuthorQueryKey,
  getPostServiceListPostsByTagQueryKey,
  getPostServiceListPostsQueryKey,
  getPostServiceSearchPostsQueryKey,
  postServiceListMyCollections,
  postServiceListPosts,
  postServiceListPostsByAuthor,
  postServiceListPostsByTag,
  postServiceSearchPosts,
  type PostListPostsResponse,
  type PostSearchPostsResponse,
  type PostServiceListMyCollectionsParams,
  type PostServiceListPostsByAuthorParams,
  type PostServiceListPostsByTagParams,
  type PostServiceListPostsParams,
  type PostServiceSearchPostsParams,
} from "@/api/generated"

type PostPage = { posts: Post[] }
type CursorPage = { nextCursorId?: string; nextCursorCreatedAt?: string }
type CursorPageParam = { cursorId?: string; cursorCreatedAt?: string }

interface UsePostInfiniteFeedOptions<TPage extends PostPage, TPageParam, TQueryKey extends QueryKey> {
  queryKey: TQueryKey
  initialPageParam: TPageParam
  enabled?: boolean
  queryFn: (pageParam: TPageParam, signal?: AbortSignal) => Promise<TPage>
  getNextPageParam: (lastPage: TPage) => TPageParam | undefined
}

export function usePostInfiniteFeed<TPage extends PostPage, TPageParam, TQueryKey extends QueryKey = QueryKey>({
  queryKey,
  initialPageParam,
  enabled = true,
  queryFn,
  getNextPageParam,
}: UsePostInfiniteFeedOptions<TPage, TPageParam, TQueryKey>) {
  const queryClient = useQueryClient()
  const query = useInfiniteQuery<TPage, unknown, InfiniteData<TPage>, TQueryKey, TPageParam>({
    queryKey,
    enabled,
    initialPageParam,
    queryFn: ({ pageParam, signal }) => queryFn(pageParam as TPageParam, signal),
    getNextPageParam,
  })

  function addPostLocal(post: Post) {
    queryClient.setQueryData<InfiniteData<TPage>>(queryKey, (old) => {
      if (!old || !old.pages.length) return old
      const firstPage = old.pages[0]
      if (firstPage.posts.some((item) => item.uid === post.uid)) return old
      const pages = [{ ...firstPage, posts: [post, ...firstPage.posts] }, ...old.pages.slice(1)]
      return { ...old, pages }
    })
  }

  function updatePostLocal(uid: string, patch: Partial<Post>) {
    queryClient.setQueryData<InfiniteData<TPage>>(queryKey, (old) => {
      if (!old) return old
      let changed = false
      const pages = old.pages.map((page) => {
        let pageChanged = false
        const posts = page.posts.map((post) => {
          if (post.uid !== uid) return post
          pageChanged = true
          changed = true
          return { ...post, ...patch }
        })
        return pageChanged ? { ...page, posts } : page
      })
      return changed ? { ...old, pages } : old
    })
  }

  function removePostLocal(uid: string) {
    queryClient.setQueryData<InfiniteData<TPage>>(queryKey, (old) => {
      if (!old) return old
      let changed = false
      const pages = old.pages.map((page) => {
        const posts = page.posts.filter((post) => post.uid !== uid)
        if (posts.length !== page.posts.length) {
          changed = true
          return { ...page, posts }
        }
        return page
      })
      return changed ? { ...old, pages } : old
    })
  }

  const posts = dedupeByUid(query.data?.pages.flatMap((page) => page.posts) ?? [])

  return { ...query, posts, queryKey, addPostLocal, updatePostLocal, removePostLocal }
}

export function buildCursorNextPageParam<TPage extends CursorPage, TPageParam extends CursorPageParam>(
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

export function useHomePostsFeed() {
  const queryKey = getPostServiceListPostsQueryKey()
  return usePostInfiniteFeed<PostListPostsResponse, PostServiceListPostsParams>({
    queryKey,
    initialPageParam: {} as PostServiceListPostsParams,
    queryFn: (pageParam, signal) => postServiceListPosts(pageParam, undefined, signal),
    getNextPageParam: (lastPage) => buildCursorNextPageParam(lastPage),
  })
}

export function useFavoritePostsFeed() {
  const queryKey = getPostServiceListMyCollectionsQueryKey()
  return usePostInfiniteFeed<PostListPostsResponse, PostServiceListMyCollectionsParams>({
    queryKey,
    initialPageParam: {} as PostServiceListMyCollectionsParams,
    queryFn: (pageParam, signal) => postServiceListMyCollections(pageParam, undefined, signal),
    getNextPageParam: (lastPage) => buildCursorNextPageParam(lastPage),
  })
}

export function useAuthorPostsFeed(uid: string) {
  const queryKey = getPostServiceListPostsByAuthorQueryKey(uid)
  return usePostInfiniteFeed<PostListPostsResponse, PostServiceListPostsByAuthorParams>({
    queryKey,
    enabled: !!uid,
    initialPageParam: {} as PostServiceListPostsByAuthorParams,
    queryFn: (pageParam, signal) => postServiceListPostsByAuthor(uid, pageParam, undefined, signal),
    getNextPageParam: (lastPage) => buildCursorNextPageParam(lastPage),
  })
}

export function useTagPostsFeed(tagName: string) {
  const queryKey = getPostServiceListPostsByTagQueryKey({ tagName })
  return usePostInfiniteFeed<PostListPostsResponse, PostServiceListPostsByTagParams>({
    queryKey,
    initialPageParam: { tagName } as PostServiceListPostsByTagParams,
    queryFn: (pageParam, signal) => postServiceListPostsByTag(pageParam, undefined, signal),
    getNextPageParam: (lastPage) => buildCursorNextPageParam(lastPage, { tagName }),
  })
}

export function useSearchPostsFeed(query: string) {
  const queryKey = getPostServiceSearchPostsQueryKey({ query })
  return usePostInfiniteFeed<PostSearchPostsResponse, PostServiceSearchPostsParams>({
    queryKey,
    enabled: !!query.trim(),
    initialPageParam: { query } as PostServiceSearchPostsParams,
    queryFn: (pageParam, signal) => postServiceSearchPosts(pageParam, undefined, signal),
    getNextPageParam: (lastPage) =>
      buildCursorNextPageParam<PostSearchPostsResponse, PostServiceSearchPostsParams>(
        lastPage,
        { query },
        { cursorScore: lastPage.nextCursorScore }
      ),
  })
}
