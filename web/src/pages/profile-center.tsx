import { ProfileCenterCard } from "@/components/profile-center-card"

export function ProfileCenter() {
  return (
    <div className="h-full w-full overflow-y-auto p-4">
      <ProfileCenterCard className="mx-auto w-full max-w-4xl" />
    </div>
  )
}
