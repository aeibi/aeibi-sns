import { Link, useLocation } from "react-router-dom"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { formatDateTime } from "@/lib/utils"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { ExternalLinkIcon, FlagIcon, MoreHorizontalIcon, Share2Icon, Trash2Icon, UserMinusIcon, UserPlusIcon } from "lucide-react"
import type { Post } from "@/types/post"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog"
import { toast } from "sonner"
import { useCopyToClipboard } from "@/hooks/use-copy-to-clipboard"

type PostHeaderProps = {
  post: Post
  isOwnPost: boolean
  onDelete: () => void
  onFollow: () => void
}

export function PostHeader({ post, isOwnPost, onDelete, onFollow }: PostHeaderProps) {
  const { search } = useLocation()
  const { copy } = useCopyToClipboard()
  const handleCopy = async () => {
    const ok = await copy(`${window.location.origin}/post/${encodeURIComponent(post.uid)}`)
    if (ok) toast.success("Share link copied to clipboard.", { position: "top-center" })
    else toast.error("Failed to copy the share link.", { position: "top-center" })
  }
  return (
    <div className="flex">
      <Link to={`/profile?uid=${encodeURIComponent(post.author.uid)}`} className="group flex items-start gap-3">
        <Avatar size="lg">
          <AvatarImage src={post.author.avatarUrl} alt={post.author.avatarUrl} />
          <AvatarFallback />
        </Avatar>
        <div className="flex flex-col">
          <span className="truncate text-base font-semibold group-hover:underline">{post.author.nickname}</span>
          <span className="text-xs text-muted-foreground">{formatDateTime(post.createdAt)}</span>
        </div>
      </Link>
      <div className="flex-1" />
      <DropdownMenu>
        <DropdownMenuTrigger render={<Button variant="ghost" size="icon-sm" />}>
          <MoreHorizontalIcon />
        </DropdownMenuTrigger>
        <DropdownMenuContent>
          <DropdownMenuItem render={<Link to={{ pathname: `post/${encodeURIComponent(post.uid)}`, search }} />}>
            <ExternalLinkIcon />
            <span>Details</span>
          </DropdownMenuItem>
          <DropdownMenuItem onClick={handleCopy}>
            <Share2Icon />
            <span>Share</span>
          </DropdownMenuItem>
          {!isOwnPost && (
            <DropdownMenuItem onClick={onFollow}>
              {post.author.isFollowing ? <UserMinusIcon /> : <UserPlusIcon />}
              <span>{post.author.isFollowing ? "Unfollow" : "Follow"}</span>
            </DropdownMenuItem>
          )}
          {!isOwnPost && (
            <DropdownMenuItem>
              <FlagIcon />
              <span>Report</span>
            </DropdownMenuItem>
          )}
          {isOwnPost && (
            <AlertDialog>
              <AlertDialogTrigger
                render={
                  <DropdownMenuItem variant="destructive" closeOnClick={false}>
                    <Trash2Icon />
                    <span>Delete</span>
                  </DropdownMenuItem>
                }
              />
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
                  <AlertDialogDescription>This action cannot be undone. This will permanently delete your post.</AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Cancel</AlertDialogCancel>
                  <AlertDialogAction variant="destructive" onClick={onDelete}>
                    Delete
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          )}
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}
