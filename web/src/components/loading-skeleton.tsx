import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/utils"

type SkeletonListProps = {
  count?: number
  className?: string
}

type SkeletonCardProps = React.ComponentProps<typeof Card>

export function SearchFormSkeleton({ className }: React.ComponentProps<"div">) {
  return <Skeleton className={cn("h-11 w-full rounded-full", className)} />
}

export function PostCommentSkeleton({ className }: React.ComponentProps<"div">) {
  return (
    <div className={cn("flex gap-2", className)}>
      <Skeleton className="size-8 rounded-full" />
      <div className="flex min-w-0 flex-1 flex-col gap-2">
        <Skeleton className="h-4 w-32" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-2/3" />
        <div className="flex gap-2">
          <Skeleton className="h-6 w-14" />
          <Skeleton className="h-6 w-14" />
          <Skeleton className="ml-auto h-6 w-6" />
        </div>
      </div>
    </div>
  )
}

export function PostReplySkeleton({ className }: React.ComponentProps<"div">) {
  return (
    <div className={cn("flex gap-2", className)}>
      <Skeleton className="size-7 rounded-full" />
      <div className="flex min-w-0 flex-1 flex-col gap-2">
        <Skeleton className="h-3.5 w-28" />
        <Skeleton className="h-3.5 w-2/3" />
      </div>
    </div>
  )
}

export function PostCardSkeleton({ className, ...props }: SkeletonCardProps) {
  return (
    <Card className={cn("gap-0", className)} {...props}>
      <CardHeader className="gap-3">
        <div className="flex items-center gap-3">
          <Skeleton className="size-10 rounded-full" />
          <div className="flex flex-1 flex-col gap-2">
            <Skeleton className="h-4 w-40" />
            <Skeleton className="h-3 w-24" />
          </div>
          <Skeleton className="size-8 rounded-md" />
        </div>
      </CardHeader>
      <CardContent className="flex flex-col gap-3">
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-4/5" />
        <Skeleton className="h-64 w-full rounded-lg" />
        <div className="flex gap-2">
          <Skeleton className="h-6 w-20" />
          <Skeleton className="h-6 w-20" />
        </div>
        <div className="flex items-center gap-2">
          <Skeleton className="h-7 w-16" />
          <Skeleton className="h-7 w-16" />
          <Skeleton className="h-7 w-16" />
        </div>
      </CardContent>
    </Card>
  )
}

export function PostListSkeleton({ count = 3, className }: SkeletonListProps) {
  return (
    <div className={cn("mx-auto flex w-full max-w-4xl flex-col gap-4 px-4 py-4", className)}>
      {Array.from({ length: count }).map((_, index) => (
        <PostCardSkeleton key={index} />
      ))}
    </div>
  )
}

export function SearchPageSkeleton({ count = 3, className }: SkeletonListProps) {
  return (
    <div className={cn("mx-auto flex w-full max-w-4xl flex-col gap-4 px-4 py-4", className)}>
      <SearchFormSkeleton />
      {Array.from({ length: count }).map((_, index) => (
        <PostCardSkeleton key={index} />
      ))}
    </div>
  )
}

export function ProfileCardSkeleton({ className, ...props }: SkeletonCardProps) {
  return (
    <Card className={cn("relative gap-0 overflow-hidden p-0", className)} {...props}>
      <CardHeader className="gap-0 p-0">
        <div className="relative h-36 w-full bg-muted/40" />
        <div className="flex flex-col gap-2 px-4 pt-14 pb-2">
          <div className="flex items-center gap-2">
            <Skeleton className="h-7 w-40" />
            <Skeleton className="h-5 w-16" />
          </div>
        </div>
      </CardHeader>
      <CardContent className="flex flex-col gap-4 pb-4">
        <div className="grid grid-cols-2 gap-2">
          <Skeleton className="h-16 w-full" />
          <Skeleton className="h-16 w-full" />
        </div>
      </CardContent>
      <Skeleton className="absolute top-24 left-4 size-20 rounded-full ring-4 ring-background" />
    </Card>
  )
}

export function ProfilePageSkeleton({ count = 2, className }: SkeletonListProps) {
  return (
    <div className={cn("mx-auto flex w-full max-w-4xl flex-col gap-4 px-4 py-4", className)}>
      <ProfileCardSkeleton />
      {Array.from({ length: count }).map((_, index) => (
        <PostCardSkeleton key={index} />
      ))}
    </div>
  )
}

export function MessageCardSkeleton({ className, ...props }: SkeletonCardProps) {
  return (
    <Card className={className} {...props}>
      <CardContent>
        <div className="flex items-start gap-3">
          <Skeleton className="size-10 rounded-full" />
          <div className="flex flex-1 flex-col gap-2">
            <Skeleton className="h-4 w-48" />
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-2/3" />
          </div>
          <Skeleton className="size-5 rounded-full" />
        </div>
      </CardContent>
    </Card>
  )
}

export function MessageListSkeleton({ count = 5, className }: SkeletonListProps) {
  return (
    <div className={cn("flex flex-col gap-2 p-1", className)}>
      {Array.from({ length: count }).map((_, index) => (
        <MessageCardSkeleton key={index} />
      ))}
    </div>
  )
}

export function RelationUserCardSkeleton({ className, ...props }: SkeletonCardProps) {
  return (
    <Card className={className} {...props}>
      <CardContent>
        <div className="flex items-start gap-3">
          <Skeleton className="size-12 rounded-full" />
          <div className="flex flex-1 flex-col gap-2">
            <div className="flex items-center gap-2">
              <Skeleton className="h-4 w-36" />
              <Skeleton className="h-4 w-16" />
            </div>
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-3/4" />
            <div className="flex gap-2">
              <Skeleton className="h-3 w-20" />
              <Skeleton className="h-3 w-20" />
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

export function RelationListSkeleton({ count = 6, className }: SkeletonListProps) {
  return (
    <div className={cn("flex flex-col gap-2 p-1", className)}>
      {Array.from({ length: count }).map((_, index) => (
        <RelationUserCardSkeleton key={index} />
      ))}
    </div>
  )
}

export function PostCommentsPreviewSkeleton({ count = 2, className }: SkeletonListProps) {
  return (
    <div className={cn("flex flex-col gap-4", className)}>
      {Array.from({ length: count }).map((_, index) => (
        <PostCommentSkeleton key={index} />
      ))}
    </div>
  )
}

export function PostRepliesSkeleton({ count = 2, className }: SkeletonListProps) {
  return (
    <div className={cn("flex flex-col gap-3", className)}>
      {Array.from({ length: count }).map((_, index) => (
        <PostReplySkeleton key={index} />
      ))}
    </div>
  )
}

export function PostDetailSkeleton({ count = 3, className }: SkeletonListProps) {
  return (
    <div className={cn("mx-auto flex w-full max-w-4xl flex-col gap-4 px-4", className)}>
      <PostCardSkeleton />
      <Card>
        <CardContent className="pt-6">
          <Skeleton className="h-24 w-full" />
        </CardContent>
      </Card>
      {Array.from({ length: count }).map((_, index) => (
        <div key={index} className="flex flex-col gap-4">
          <PostCommentSkeleton />
          <Skeleton className="ml-8 h-px w-full" />
        </div>
      ))}
    </div>
  )
}

export function ProfileCenterCardSkeleton({ className, ...props }: SkeletonCardProps) {
  return (
    <Card className={className} {...props}>
      <CardHeader>
        <Skeleton className="h-6 w-36" />
      </CardHeader>
      <CardContent>
        <div className="grid gap-8">
          <div className="grid gap-6">
            <Skeleton className="size-20 rounded-full" />
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-10 w-full" />
            <div className="flex gap-2">
              <Skeleton className="h-9 w-28" />
              <Skeleton className="h-9 w-20" />
            </div>
          </div>
          <Skeleton className="h-px w-full" />
          <div className="grid gap-6">
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-9 w-32" />
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
