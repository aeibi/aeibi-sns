import { SearchForm } from "@/components/search-form"
import { useSearchParams } from "react-router-dom"
import { useUserServiceGetMe } from "@/api/generated"
import { PostListSkeleton } from "@/components/loading-skeleton"
import { PostFeedList } from "@/components/post-feed-list"
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/components/ui/empty"
import { useTagPostsFeed } from "@/hooks/use-post-infinite-feed"

export function Tag() {
  const { data: userData } = useUserServiceGetMe()

  const [searchParams] = useSearchParams()
  const tagName = searchParams.get("tag") ?? ""
  const { posts, fetchNextPage, isFetchingNextPage, hasNextPage, isPending, updatePostLocal, removePostLocal } = useTagPostsFeed(tagName)

  if (isPending && !posts.length) return <TagSkeleton tagName={tagName} />
  if (!posts.length) return <TagEmpty tagName={tagName} />
  return (
    <PostFeedList
      posts={posts}
      user={userData?.user}
      headerKey="search-form"
      header={<SearchForm className="w-full" searchText={tagName} />}
      hasNextPage={hasNextPage}
      isFetchingNextPage={isFetchingNextPage}
      onLoadMore={fetchNextPage}
      onRemovePost={removePostLocal}
      onUpdatePost={updatePostLocal}
    />
  )
}

function TagSkeleton({ tagName }: { tagName: string }) {
  return (
    <div className="h-full w-full overflow-y-auto">
      <SearchForm className="w-full" searchText={tagName} />
      <PostListSkeleton count={3} />
    </div>
  )
}

function TagEmpty({ tagName }: { tagName: string }) {
  return (
    <div className="h-full w-full p-4">
      <div className="mx-auto flex h-full w-full max-w-4xl flex-col gap-4 px-4 py-4">
        <SearchForm className="w-full" searchText={tagName} />
        <Empty className="h-full border">
          <EmptyHeader>
            <EmptyTitle>{tagName.trim() ? "No Tagged Posts" : "Tag Not Specified"}</EmptyTitle>
            <EmptyDescription>
              {tagName.trim() ? "There are no posts under this tag yet." : "Choose a tag to view related posts."}
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      </div>
    </div>
  )
}
