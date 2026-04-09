import { Card, CardContent } from "@/components/ui/card"
import { InputGroup, InputGroupAddon, InputGroupInput } from "@/components/ui/input-group"
import { SearchIcon } from "lucide-react"
import { useEffect, useState } from "react"

type RelationSearchCardProps = {
  query: string
  onQueryChange: (query: string) => void
}

export function RelationSearchCard({ query, onQueryChange }: RelationSearchCardProps) {
  const [text, setText] = useState(query)
  const [debouncedText, setDebouncedText] = useState(query)

  useEffect(() => {
    setText(query)
  }, [query])

  useEffect(() => {
    const id = window.setTimeout(() => {
      setDebouncedText(text)
    }, 200)
    return () => window.clearTimeout(id)
  }, [text])

  useEffect(() => {
    if (debouncedText !== query) {
      onQueryChange(debouncedText)
    }
  }, [debouncedText, onQueryChange, query])

  return (
    <Card>
      <CardContent className="flex justify-center">
        <InputGroup>
          <InputGroupInput value={text} placeholder="Search users" onChange={(event) => setText(event.target.value)} />
          <InputGroupAddon>
            <SearchIcon />
          </InputGroupAddon>
        </InputGroup>
      </CardContent>
    </Card>
  )
}
