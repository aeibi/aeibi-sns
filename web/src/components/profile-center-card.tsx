import { useUserServiceChangePassword, useUserServiceGetMe, useUserServiceUpdateMe, type UserUpdateMeUser } from "@/api/generated"
import { ProfileCenterCardSkeleton } from "@/components/loading-skeleton"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from "@/components/ui/empty"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { useUploadFiles } from "@/hooks/use-upload-files"
import { cn } from "@/lib/utils"
import { useState, type ChangeEvent, type SubmitEvent } from "react"
import { Separator } from "@/components/ui/separator"
import { Label } from "@/components/ui/label"
import { toast } from "sonner"

export function ProfileCenterCard({ className, ...props }: React.ComponentProps<typeof Card>) {
  const { data: userData, refetch, isPending: isUserPending } = useUserServiceGetMe()
  const { mutate: updateMe, isPending: isUpdating, error: updateError } = useUserServiceUpdateMe()
  const { mutate: changePassword, isPending: isChangingPassword, error: changePasswordError } = useUserServiceChangePassword()
  const { mutate: uploadFiles, isPending: isUploadingAvatar, error: uploadError } = useUploadFiles()

  const [profile, setProfile] = useState<UserUpdateMeUser>({})
  const avatarUrl = profile.avatarUrl ?? userData?.user.avatarUrl ?? ""
  const nickname = profile.nickname ?? userData?.user.nickname ?? ""
  const email = profile.email ?? userData?.user.email ?? ""

  const handleAvatarChange = (event: ChangeEvent<HTMLInputElement>) => {
    event.preventDefault()
    const files = Array.from(event.target.files ?? []).filter((file) => file.type.startsWith("image/"))
    event.currentTarget.value = ""
    if (!files.length) return
    const file = files[0]
    if (!file) return
    uploadFiles(
      { files: [file] },
      {
        onSuccess: (data) => {
          setProfile((current) => ({ ...current, avatarUrl: data.urls[0] }))
        },
      }
    )
  }

  const onResetProfileForm = () => {
    setProfile({})
  }

  const handleProfileSubmit = (event: SubmitEvent<HTMLFormElement>) => {
    event.preventDefault()
    updateMe(
      { data: profile },
      {
        onSuccess: () => {
          toast.success("Profile updated", { position: "top-center" })
          void refetch().finally(() => setProfile({}))
        },
      }
    )
  }

  const handleChangePassword = (event: SubmitEvent<HTMLFormElement>) => {
    event.preventDefault()
    const form = event.currentTarget
    const formData = new FormData(form)
    const oldPassword = formData.get("old-password")?.toString()
    const newPassword = formData.get("new-password")?.toString()
    const confirmPassword = formData.get("confirm-password")?.toString()
    if (newPassword !== confirmPassword) {
      toast.error("New password and confirm password do not match", { position: "top-center" })
      return
    }
    if (!oldPassword || !newPassword) {
      toast.error("Please fill in all password fields", { position: "top-center" })
      return
    }
    if (newPassword.length < 8) {
      toast.error("New password must be at least 8 characters long", { position: "top-center" })
      return
    }
    changePassword(
      { data: { oldPassword, newPassword } },
      {
        onSuccess: () => {
          form.reset()
          toast.success("Password updated", { position: "top-center" })
        },
      }
    )
  }

  if (isUserPending) return <ProfileCenterCardSkeleton className={className} {...props} />
  if (!userData) {
    return (
      <Card className={cn("h-full", className)} {...props}>
        <CardContent className="h-full">
          <Empty className="h-full">
            <EmptyHeader>
              <EmptyTitle>Profile Unavailable</EmptyTitle>
              <EmptyDescription>Please log in to access profile settings.</EmptyDescription>
            </EmptyHeader>
          </Empty>
        </CardContent>
      </Card>
    )
  }
  return (
    <Card className={className} {...props}>
      <CardHeader>
        <CardTitle>Profile Center</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid gap-8">
          <form onSubmit={handleProfileSubmit}>
            <FieldGroup>
              <Field>
                <input
                  id="avatar"
                  type="file"
                  accept="image/*"
                  className="hidden"
                  disabled={isUploadingAvatar}
                  onChange={handleAvatarChange}
                />
                <div>
                  <Button className="size-20 rounded-full" render={<Label htmlFor="avatar" />}>
                    <Avatar className="size-20">
                      <AvatarImage src={avatarUrl} />
                      <AvatarFallback />
                    </Avatar>
                  </Button>
                </div>
                {uploadError && <FieldError>{uploadError.message}</FieldError>}
              </Field>
              <Field>
                <FieldLabel htmlFor="profile-nickname">Nickname</FieldLabel>
                <Input
                  id="profile-nickname"
                  name="profile-nickname"
                  placeholder="aeibi"
                  value={nickname}
                  onChange={(event) => {
                    const nickname = event.currentTarget.value
                    setProfile((current) => ({ ...current, nickname }))
                  }}
                />
              </Field>
              <Field>
                <FieldLabel htmlFor="profile-email">Email</FieldLabel>
                <Input
                  id="profile-email"
                  name="profile-email"
                  type="email"
                  placeholder="m@example.com"
                  value={email}
                  onChange={(event) => {
                    const email = event.currentTarget.value
                    setProfile((current) => ({ ...current, email }))
                  }}
                />
              </Field>
              <Field>
                <div className="flex gap-2">
                  <Button type="submit" disabled={isUpdating}>
                    {isUpdating ? "Saving..." : "Save profile"}
                  </Button>
                  <Button type="button" variant="outline" onClick={onResetProfileForm}>
                    Reset
                  </Button>
                </div>
                {updateError && <FieldError>{updateError.message}</FieldError>}
              </Field>
            </FieldGroup>
          </form>
          <Separator />
          <form onSubmit={handleChangePassword}>
            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="old-password">Current password</FieldLabel>
                <Input id="old-password" name="old-password" type="password" autoComplete="current-password" required />
              </Field>
              <Field>
                <FieldLabel htmlFor="new-password">New password</FieldLabel>
                <Input id="new-password" name="new-password" type="password" autoComplete="new-password" required />
              </Field>
              <Field>
                <FieldLabel htmlFor="confirm-password">Confirm new password</FieldLabel>
                <Input id="confirm-password" name="confirm-password" type="password" autoComplete="new-password" required />
              </Field>
              <Field>
                <Button type="submit" disabled={isChangingPassword}>
                  {isChangingPassword ? "Updating..." : "Update password"}
                </Button>
                {changePasswordError && <FieldError>{changePasswordError.message}</FieldError>}
              </Field>
            </FieldGroup>
          </form>
        </div>
      </CardContent>
    </Card>
  )
}
