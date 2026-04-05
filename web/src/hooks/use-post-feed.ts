import { useQueryClient } from "@tanstack/react-query"
import { getPostServiceGetPostQueryKey, type PostGetPostResponse, usePostServiceGetPost } from "@/api/generated"
import type { Post } from "@/types/post"

export function usePostFeed(uid: string) {
  const queryClient = useQueryClient()
  const query = usePostServiceGetPost(uid)
  const queryKey = getPostServiceGetPostQueryKey(uid)

  function updatePostLocal(postUid: string, patch: Partial<Post>) {
    queryClient.setQueryData<PostGetPostResponse>(queryKey, (old) => {
      if (!old || old.post.uid !== postUid) return old
      return { ...old, post: { ...old.post, ...patch } }
    })
  }

  return { ...query, post: query.data?.post, updatePostLocal }
}
