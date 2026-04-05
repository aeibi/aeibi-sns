import { Card, CardContent } from "@/components/ui/card"
import { InputGroup, InputGroupAddon, InputGroupInput } from "@/components/ui/input-group"
import { SearchIcon } from "lucide-react"

type RelationSearchCardProps = {
  query: string
  onQueryChange: (query: string) => void
}

export function RelationSearchCard({ query, onQueryChange }: RelationSearchCardProps) {
  return (
    <Card>
      <CardContent className="flex justify-center">
        <InputGroup>
          <InputGroupInput value={query} placeholder="Search users" onChange={(event) => onQueryChange(event.target.value)} />
          <InputGroupAddon>
            <SearchIcon />
          </InputGroupAddon>
        </InputGroup>
      </CardContent>
    </Card>
  )
}
