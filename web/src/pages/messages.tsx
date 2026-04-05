import { useInfiniteQuery } from "@tanstack/react-query"
import { MessageCategorySidenav, type MessageCategory } from "@/components/message-category-sidenav"
import { MessageCommentCard } from "@/components/message-comment-card"
import { MessageFollowCard } from "@/components/message-follow-card"
import { MessageListSkeleton } from "@/components/loading-skeleton"
import { MessageStatusTabs, type MessageStatus } from "@/components/message-status-tabs"
import { VirtualList } from "@/components/virtual-list"
import {
  getMessageServiceListCommentInboxMessagesQueryKey,
  getMessageServiceListFollowInboxMessagesQueryKey,
  useMessageServiceMarkAllInboxMessagesRead,
  messageServiceListCommentInboxMessages,
  messageServiceListFollowInboxMessages,
  type MessageServiceListCommentInboxMessagesParams,
  type MessageServiceListFollowInboxMessagesParams,
} from "@/api/generated"
import { useSearchParams } from "react-router-dom"
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/components/ui/empty"

const InboxMessageReadFilter = {
  UNSPECIFIED: 0,
  UNREAD: 1,
  READ: 2,
} as const

export function Messages() {
  const [searchParams, setSearchParams] = useSearchParams()
  const category = searchParams.get("category") === "comment" ? "comment" : "follow"
  const status = searchParams.get("status") === "all" ? "all" : "unread"
  const { mutate: markAllInboxMessagesRead } = useMessageServiceMarkAllInboxMessagesRead()

  const readFilter = status === "unread" ? InboxMessageReadFilter.UNREAD : InboxMessageReadFilter.UNSPECIFIED

  const {
    data: followData,
    fetchNextPage: fetchFollowNextPage,
    isFetchingNextPage: isFetchingFollowNextPage,
    hasNextPage: hasFollowNextPage,
    isPending: isFollowPending,
    refetch: refetchFollowMessages,
  } = useInfiniteQuery({
    queryKey: getMessageServiceListFollowInboxMessagesQueryKey({ readFilter }),
    initialPageParam: { readFilter } as MessageServiceListFollowInboxMessagesParams,
    queryFn: ({ pageParam, signal }) => messageServiceListFollowInboxMessages(pageParam, undefined, signal),
    getNextPageParam: (lastPage) => {
      if (!lastPage.nextCursorId || !lastPage.nextCursorCreatedAt) return
      return {
        cursorId: lastPage.nextCursorId,
        cursorCreatedAt: lastPage.nextCursorCreatedAt,
        readFilter,
      }
    },
    enabled: category === "follow",
  })

  const {
    data: commentData,
    fetchNextPage: fetchCommentNextPage,
    isFetchingNextPage: isFetchingCommentNextPage,
    hasNextPage: hasCommentNextPage,
    isPending: isCommentPending,
    refetch: refetchCommentMessages,
  } = useInfiniteQuery({
    queryKey: getMessageServiceListCommentInboxMessagesQueryKey({ readFilter }),
    initialPageParam: { readFilter } as MessageServiceListCommentInboxMessagesParams,
    queryFn: ({ pageParam, signal }) => messageServiceListCommentInboxMessages(pageParam, undefined, signal),
    getNextPageParam: (lastPage) => {
      if (!lastPage.nextCursorId || !lastPage.nextCursorCreatedAt) return
      return {
        cursorId: lastPage.nextCursorId,
        cursorCreatedAt: lastPage.nextCursorCreatedAt,
        readFilter,
      }
    },
    enabled: category === "comment",
  })

  const followMessages = followData?.pages.flatMap((page) => page.messages) ?? []
  const commentMessages = commentData?.pages.flatMap((page) => page.messages) ?? []
  const isFollowCategory = category === "follow"
  const activeMessages = isFollowCategory ? followMessages : commentMessages
  const fetchNextPage = isFollowCategory ? fetchFollowNextPage : fetchCommentNextPage
  const isFetchingNextPage = isFollowCategory ? isFetchingFollowNextPage : isFetchingCommentNextPage
  const hasNextPage = isFollowCategory ? hasFollowNextPage : hasCommentNextPage
  const isPending = isFollowCategory ? isFollowPending : isCommentPending
  const refetchMessages = isFollowCategory ? refetchFollowMessages : refetchCommentMessages

  const handleCategoryChange = (nextCategory: MessageCategory) => {
    const nextSearchParams = new URLSearchParams(searchParams)
    nextSearchParams.set("category", nextCategory)
    setSearchParams(nextSearchParams)
  }

  const handleStatusChange = (nextStatus: MessageStatus) => {
    const nextSearchParams = new URLSearchParams(searchParams)
    nextSearchParams.set("status", nextStatus)
    setSearchParams(nextSearchParams)
  }

  const handleMarkAllAsRead = () => {
    markAllInboxMessagesRead()
    const nextSearchParams = new URLSearchParams(searchParams)
    nextSearchParams.set("status", "all")
    setSearchParams(nextSearchParams)
    void refetchMessages()
  }

  return (
    <div className="flex h-full min-h-0 justify-center gap-4 p-4">
      <MessageCategorySidenav selectedCategory={category} onCategoryChange={handleCategoryChange} className="h-full w-56" />
      <div className="flex min-h-0 w-full max-w-4xl flex-col gap-4">
        <MessageStatusTabs selectedStatus={status} onStatusChange={handleStatusChange} onMarkAllAsRead={handleMarkAllAsRead} />
        {!activeMessages.length && !isPending ? (
          <div className="min-h-0 flex-1">
            <Empty className="h-full border">
              <EmptyHeader>
                <EmptyTitle>No Messages</EmptyTitle>
                <EmptyDescription>Your inbox is clear for now.</EmptyDescription>
              </EmptyHeader>
            </Empty>
          </div>
        ) : (
          <>
            {!activeMessages.length && isPending && (
              <div className="min-h-0 flex-1 overflow-y-auto">
                <MessageListSkeleton />
              </div>
            )}
            {!!activeMessages.length &&
              (isFollowCategory ? (
                <VirtualList
                  key={`follow-${status}`}
                  items={followMessages}
                  getItemKey={(message) => message.uid}
                  hasNextPage={hasNextPage}
                  isFetchingNextPage={isFetchingNextPage}
                  onLoadMore={fetchNextPage}
                  estimateSize={() => 150}
                  gap={8}
                  paddingStart={4}
                  paddingEnd={4}
                  className="min-h-0 flex-1 overflow-y-auto"
                  innerClassName="w-full"
                  renderItem={(message) => <MessageFollowCard message={message} />}
                />
              ) : (
                <VirtualList
                  key={`comment-${status}`}
                  items={commentMessages}
                  getItemKey={(message) => message.uid}
                  hasNextPage={hasNextPage}
                  isFetchingNextPage={isFetchingNextPage}
                  onLoadMore={fetchNextPage}
                  estimateSize={() => 150}
                  gap={8}
                  paddingStart={4}
                  paddingEnd={4}
                  className="min-h-0 flex-1 overflow-y-auto"
                  innerClassName="w-full"
                  renderItem={(message) => <MessageCommentCard message={message} />}
                />
              ))}
          </>
        )}
      </div>
    </div>
  )
}
