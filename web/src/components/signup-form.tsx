import { useUserServiceCreateUser } from "@/api/generated"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Field, FieldDescription, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { useState, type SubmitEvent } from "react"
import { Link, useNavigate } from "react-router-dom"

export function SignupForm({ ...props }: React.ComponentProps<typeof Card>) {
  const navigate = useNavigate()
  const [formError, setFormError] = useState<string | null>(null)
  const { mutate, isPending, error } = useUserServiceCreateUser()

  const handleSubmit = (event: SubmitEvent<HTMLFormElement>) => {
    event.preventDefault()
    setFormError(null)
    const formData = new FormData(event.currentTarget)
    const email = String(formData.get("email")).trim()
    const nickname = String(formData.get("nickname")).trim()
    const password = String(formData.get("password"))
    const confirmPassword = String(formData.get("confirm-password"))
    if (password !== confirmPassword) {
      setFormError("Passwords do not match.")
      return
    }
    mutate({ data: { username: email, email, password, nickname } }, { onSuccess: () => navigate("/login") })
  }

  return (
    <Card {...props}>
      <CardHeader>
        <CardTitle>Create an account</CardTitle>
        <CardDescription>Enter your information below to create your account</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit}>
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="nickname">Nickname</FieldLabel>
              <Input id="nickname" name="nickname" type="text" placeholder="yoclo" required />
            </Field>
            <Field>
              <FieldLabel htmlFor="email">Email</FieldLabel>
              <Input id="email" name="email" type="email" placeholder="m@example.com" required />
              <FieldDescription>We&apos;ll use this to contact you. We will not share your email with anyone else.</FieldDescription>
            </Field>
            <Field>
              <FieldLabel htmlFor="password">Password</FieldLabel>
              <Input id="password" name="password" type="password" required />
              <FieldDescription>Must be at least 8 characters long.</FieldDescription>
            </Field>
            <Field>
              <FieldLabel htmlFor="confirm-password">Confirm Password</FieldLabel>
              <Input id="confirm-password" name="confirm-password" type="password" required />
              <FieldDescription>Please confirm your password.</FieldDescription>
            </Field>
            <FieldGroup>
              <Field>
                <Button type="submit" disabled={isPending}>
                  {isPending ? "Creating..." : "Create Account"}
                </Button>
                {formError && <FieldError>{formError}</FieldError>}
                {error?.message && <FieldError>{error.message}</FieldError>}
                <FieldDescription className="px-6 text-center">
                  Already have an account? <Link to="/login">Sign in</Link>
                </FieldDescription>
              </Field>
            </FieldGroup>
          </FieldGroup>
        </form>
      </CardContent>
    </Card>
  )
}
