import { useCallback, useState } from "react"

export function useCopyToClipboard(timeout = 1500) {
  const [copied, setCopied] = useState(false)

  const copy = useCallback(
    async (text: string) => {
      try {
        await navigator.clipboard.writeText(text)
        setCopied(true)

        window.setTimeout(() => {
          setCopied(false)
        }, timeout)

        return true
      } catch {
        setCopied(false)
        return false
      }
    },
    [timeout]
  )

  return { copied, copy }
}
