import { token } from "@/api/client"
import { useUserServiceLogin } from "@/api/generated"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Field, FieldDescription, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import type { SubmitEvent } from "react"
import { Link, useNavigate } from "react-router-dom"

export function LoginForm({ ...props }: React.ComponentProps<typeof Card>) {
  const navigate = useNavigate()
  const { mutate, isPending, error } = useUserServiceLogin()

  const handleSubmit = (event: SubmitEvent<HTMLFormElement>) => {
    event.preventDefault()
    const formData = new FormData(event.currentTarget)
    const account = String(formData.get("email")).trim()
    const password = String(formData.get("password"))
    mutate(
      { data: { account, password } },
      {
        onSuccess: (data) => {
          token.set(data.tokens)
          navigate("/")
        },
      }
    )
  }

  return (
    <Card {...props}>
      <CardHeader>
        <CardTitle>Login to your account</CardTitle>
        <CardDescription>Enter your email below to login to your account</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit}>
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="email">Email</FieldLabel>
              <Input id="email" name="email" type="email" placeholder="m@example.com" required />
            </Field>
            <Field>
              <FieldLabel htmlFor="password">Password</FieldLabel>
              <Input id="password" name="password" type="password" required />
            </Field>
            <Field>
              <Button type="submit" disabled={isPending}>
                {isPending ? "Logging in..." : "Login"}
              </Button>
              {error?.message && <FieldError>{error.message}</FieldError>}
              <FieldDescription className="text-center">
                Don&apos;t have an account? <Link to="/signup">Sign up</Link>
              </FieldDescription>
              <FieldDescription className="text-center">
                <Link to="/">Continue without logging in</Link>
              </FieldDescription>
            </Field>
          </FieldGroup>
        </form>
      </CardContent>
    </Card>
  )
}
