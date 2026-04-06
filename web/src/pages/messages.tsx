import { useInfiniteQuery } from "@tanstack/react-query"
import { MessageCategorySidenav, type MessageCategory } from "@/components/message-category-sidenav"
import { MessageCommentCard } from "@/components/message-comment-card"
import { MessageFollowCard } from "@/components/message-follow-card"
import { MessageStatusTabs, type MessageStatus } from "@/components/message-status-tabs"
import { VirtualList } from "@/components/virtual-list"
import type { CommentMessage, FollowMessage, InboxMessage } from "@/types/message"
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
    refetch: refetchFollowMessages,
  } = useInfiniteQuery({
    queryKey: getMessageServiceListFollowInboxMessagesQueryKey({ readFilter }),
    initialPageParam: { readFilter } as MessageServiceListFollowInboxMessagesParams,
    queryFn: ({ pageParam, signal }) => messageServiceListFollowInboxMessages(pageParam, undefined, signal),
    getNextPageParam: (lastPage) => {
      if (!lastPage.nextPageToken) return
      return {
        pageToken: lastPage.nextPageToken,
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
    refetch: refetchCommentMessages,
  } = useInfiniteQuery({
    queryKey: getMessageServiceListCommentInboxMessagesQueryKey({ readFilter }),
    initialPageParam: { readFilter } as MessageServiceListCommentInboxMessagesParams,
    queryFn: ({ pageParam, signal }) => messageServiceListCommentInboxMessages(pageParam, undefined, signal),
    getNextPageParam: (lastPage) => {
      if (!lastPage.nextPageToken) return
      return {
        pageToken: lastPage.nextPageToken,
        readFilter,
      }
    },
    enabled: category === "comment",
  })

  const followMessages: FollowMessage[] = followData?.pages.flatMap((page) => page.messages) ?? []
  const commentMessages: CommentMessage[] = commentData?.pages.flatMap((page) => page.messages) ?? []
  const isFollowCategory = category === "follow"
  const activeMessages: InboxMessage[] = isFollowCategory ? followMessages : commentMessages
  const fetchNextPage = isFollowCategory ? fetchFollowNextPage : fetchCommentNextPage
  const isFetchingNextPage = isFollowCategory ? isFetchingFollowNextPage : isFetchingCommentNextPage
  const hasNextPage = isFollowCategory ? hasFollowNextPage : hasCommentNextPage
  const refetchMessages = isFollowCategory ? refetchFollowMessages : refetchCommentMessages

  const updateSearchParam = (key: "category" | "status", value: MessageCategory | MessageStatus) => {
    const nextSearchParams = new URLSearchParams(searchParams)
    nextSearchParams.set(key, value)
    setSearchParams(nextSearchParams)
  }

  const handleCategoryChange = (nextCategory: MessageCategory) => {
    updateSearchParam("category", nextCategory)
  }

  const handleStatusChange = (nextStatus: MessageStatus) => {
    updateSearchParam("status", nextStatus)
  }

  const handleMarkAllAsRead = () => {
    markAllInboxMessagesRead(undefined, {
      onSuccess: () => {
        updateSearchParam("status", "all")
        void refetchMessages()
      },
    })
  }

  return (
    <div className="flex h-full min-h-0 justify-center gap-4 p-4">
      <MessageCategorySidenav selectedCategory={category} onCategoryChange={handleCategoryChange} className="h-full w-56" />
      <div className="flex min-h-0 w-full max-w-4xl flex-col gap-4">
        <MessageStatusTabs selectedStatus={status} onStatusChange={handleStatusChange} onMarkAllAsRead={handleMarkAllAsRead} />
        <VirtualList
          key={`${category}-${status}`}
          items={activeMessages}
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
          renderItem={(message) =>
            isFollowCategory ? (
              <MessageFollowCard message={message as FollowMessage} />
            ) : (
              <MessageCommentCard message={message as CommentMessage} />
            )
          }
        />
      </div>
    </div>
  )
}
