import { useInfiniteQuery, useQueryClient, type InfiniteData, type QueryKey } from "@tanstack/react-query"
import { dedupeByUid } from "@/lib/utils"
import type { Post } from "@/types/post"
import {
  getPostServiceListMyCollectionsQueryKey,
  getPostServiceListPostsQueryKey,
  postServiceListMyCollections,
  postServiceListPosts,
  type PostListPostsResponse,
  type PostServiceListPostsParams,
} from "@/api/generated"

interface UsePostInfiniteFeedOptions {
  queryKey: QueryKey
  initialPageParam: PostServiceListPostsParams
  enabled?: boolean
  queryFn: (pageParam: PostServiceListPostsParams, signal?: AbortSignal) => Promise<PostListPostsResponse>
  getNextPageParam: (lastPage: PostListPostsResponse) => PostServiceListPostsParams | undefined
}

export function usePostInfiniteFeed({ queryKey, initialPageParam, enabled = true, queryFn, getNextPageParam }: UsePostInfiniteFeedOptions) {
  const queryClient = useQueryClient()
  const query = useInfiniteQuery<PostListPostsResponse, unknown, InfiniteData<PostListPostsResponse>, QueryKey, PostServiceListPostsParams>(
    {
      queryKey,
      enabled,
      initialPageParam,
      queryFn: ({ pageParam, signal }) => queryFn(pageParam, signal),
      getNextPageParam,
    }
  )

  function addPostLocal(post: Post) {
    queryClient.setQueryData<InfiniteData<PostListPostsResponse>>(queryKey, (old) => {
      if (!old || !old.pages.length) return old
      const firstPage = old.pages[0]
      if (firstPage.posts.some((item) => item.uid === post.uid)) return old
      const pages = [{ ...firstPage, posts: [post, ...firstPage.posts] }, ...old.pages.slice(1)]
      return { ...old, pages }
    })
  }

  function updatePostLocal(uid: string, patch: Partial<Post>) {
    queryClient.setQueryData<InfiniteData<PostListPostsResponse>>(queryKey, (old) => {
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
    queryClient.setQueryData<InfiniteData<PostListPostsResponse>>(queryKey, (old) => {
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

export function useFavoritePostsFeed() {
  const queryKey = getPostServiceListMyCollectionsQueryKey()
  return usePostInfiniteFeed({
    queryKey,
    initialPageParam: {},
    queryFn: (pageParam, signal) => postServiceListMyCollections(pageParam, undefined, signal),
    getNextPageParam: (lastPage) => {
      if (!lastPage.nextPageToken) return
      return {
        pageToken: lastPage.nextPageToken,
      }
    },
  })
}

export function usePostsFeed(params: PostServiceListPostsParams) {
  const queryKey = getPostServiceListPostsQueryKey(params)
  return usePostInfiniteFeed({
    queryKey,
    initialPageParam: params,
    queryFn: (pageParam, signal) => postServiceListPosts(pageParam, undefined, signal),
    getNextPageParam: (lastPage) => {
      if (!lastPage.nextPageToken) return
      return {
        ...(params ?? {}),
        pageToken: lastPage.nextPageToken,
      }
    },
  })
}
