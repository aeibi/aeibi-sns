import { useQueryClient, type InfiniteData, type QueryKey } from "@tanstack/react-query"
import type { Comment } from "@/types/comment"

type CommentPage = { comments: Comment[] }

export function useCommentListLocalActions<TPage extends CommentPage>(queryKey: QueryKey) {
  const queryClient = useQueryClient()

  function addCommentLocal(comment: Comment) {
    queryClient.setQueryData<InfiniteData<TPage>>(queryKey, (old) => {
      if (!old || !Array.isArray(old.pages) || old.pages.length === 0) return old
      const firstPage = old.pages[0]
      if (firstPage.comments.some((item) => item.uid === comment.uid)) return old
      const pages = [{ ...firstPage, comments: [comment, ...firstPage.comments] }, ...old.pages.slice(1)]
      return { ...old, pages }
    })
  }

  function updateCommentLocal(uid: string, patch: Partial<Comment>) {
    queryClient.setQueryData<InfiniteData<TPage>>(queryKey, (old) => {
      if (!old || !Array.isArray(old.pages)) return old
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
      if (!old || !Array.isArray(old.pages)) return old
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

  return { addCommentLocal, updateCommentLocal, removeCommentLocal }
}
