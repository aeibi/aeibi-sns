import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

const dateTimeFormatter = new Intl.DateTimeFormat("en", {
  dateStyle: "medium",
  timeStyle: "short",
})
export const formatDateTime = (input: string) => dateTimeFormatter.format(new Date(Number(input) * 1000))

const compactNumber = new Intl.NumberFormat("en", {
  notation: "compact",
  maximumFractionDigits: 1,
})
export function formatCount(value: number) {
  return compactNumber.format(value)
}

export function dedupeByUid<T extends { uid: string }>(items: T[]) {
  const seen = new Set<string>()
  return items.filter((item) => {
    if (seen.has(item.uid)) return false
    seen.add(item.uid)
    return true
  })
}
