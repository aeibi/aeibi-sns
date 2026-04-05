import { usePostServiceSuggestTagsByPrefix, useUserServiceSuggestUsersByPrefix } from "@/api/generated"
import { SearchIcon } from "lucide-react"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { Command, CommandDialog, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command"
import { useEffect, useState } from "react"
import { useNavigate } from "react-router-dom"
import { keepPreviousData } from "@tanstack/react-query"

type SearchFormProps = React.ComponentProps<typeof Button> & {
  searchText?: string
}

export function SearchForm({ className, searchText, ...props }: SearchFormProps) {
  const navigate = useNavigate()
  const [open, setOpen] = useState(false)
  const [text, setText] = useState("")

  const [debouncedText, setDebouncedText] = useState("")
  useEffect(() => {
    const id = window.setTimeout(() => {
      setDebouncedText(text)
    }, 200)
    return () => window.clearTimeout(id)
  }, [text])

  const shouldSuggest = open && debouncedText.length > 0

  const { data: usersData } = useUserServiceSuggestUsersByPrefix(
    { prefix: debouncedText },
    {
      query: {
        enabled: shouldSuggest,
        placeholderData: keepPreviousData,
        staleTime: 15_000,
      },
    }
  )

  const { data: tagsData } = usePostServiceSuggestTagsByPrefix(
    { prefix: debouncedText },
    {
      query: {
        enabled: shouldSuggest,
        placeholderData: keepPreviousData,
        staleTime: 15_000,
      },
    }
  )

  const handleSearch = () => {
    const query = text.trim()
    if (!query) return
    navigate(`/search?query=${encodeURIComponent(query)}`)
    setOpen(false)
  }

  const handleOpenChange = (nextOpen: boolean) => {
    setOpen(nextOpen)
    if (nextOpen) setText("")
    if (searchText && nextOpen) setText(searchText)
  }
  return (
    <>
      <Button
        onClick={() => handleOpenChange(true)}
        variant="outline"
        size="lg"
        className={cn("justify-start font-normal", className)}
        {...props}
      >
        <SearchIcon />
        <span
          className={cn("w-full truncate text-start", searchText ? "text-foreground" : "hidden w-80 text-muted-foreground md:inline")}
          title={searchText}
        >
          {searchText ? searchText : "Type to search..."}
        </span>
      </Button>
      <CommandDialog open={open} onOpenChange={handleOpenChange} className="top-8 sm:max-w-[calc(100%-2rem)]">
        <Command shouldFilter={false}>
          <CommandInput placeholder="Type to search..." autoFocus value={text} onValueChange={setText} />
          <CommandList>
            <CommandEmpty>No results found.</CommandEmpty>
            <CommandGroup>
              <CommandItem value={`search:${text}`} onSelect={handleSearch}>
                <span className="truncate">Search posts for &quot;{text}&quot;</span>
              </CommandItem>
            </CommandGroup>
            {shouldSuggest && !!usersData?.users.length && (
              <CommandGroup heading="Users">
                {usersData.users.map((user) => (
                  <CommandItem
                    key={user.uid}
                    value={`user:${user.uid}`}
                    onSelect={() => {
                      navigate(`/profile?uid=${encodeURIComponent(user.uid)}`)
                      setOpen(false)
                    }}
                  >
                    {user.nickname}
                  </CommandItem>
                ))}
              </CommandGroup>
            )}
            {shouldSuggest && !!tagsData?.tags.length && (
              <CommandGroup heading="Tags">
                {tagsData.tags.map((tag) => (
                  <CommandItem
                    key={tag.name}
                    value={`tag:${tag.name}`}
                    onSelect={() => {
                      navigate(`/tag?tag=${encodeURIComponent(tag.name)}`)
                      setOpen(false)
                    }}
                  >
                    {tag.name}
                  </CommandItem>
                ))}
              </CommandGroup>
            )}
          </CommandList>
        </Command>
      </CommandDialog>
    </>
  )
}
