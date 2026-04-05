import { Button } from "@/components/ui/button"
import { SunIcon, MoonIcon } from "lucide-react"
import { Link } from "react-router-dom"
import { useTheme } from "@/components/theme-provider"

export function HeaderAction({
  items,
}: {
  items: {
    icon: React.ReactNode
    url: string
  }[]
}) {
  const { theme, setTheme } = useTheme()
  const isDark = theme === "dark"
  return (
    <>
      <Button variant="ghost" size="icon-lg" onClick={() => setTheme(isDark ? "light" : "dark")}>
        {isDark ? <SunIcon /> : <MoonIcon />}
      </Button>
      {items.map((item) => (
        <Button variant="ghost" size="icon-lg" render={<Link to={item.url} />}>
          {item.icon}
        </Button>
      ))}
    </>
  )
}
