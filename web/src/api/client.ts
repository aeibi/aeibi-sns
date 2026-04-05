import axios from "axios"
import createAuthRefreshInterceptor from "axios-auth-refresh"
import { userServiceRefreshToken, type UserTokenPair } from "./generated"

const baseURL = import.meta.env.VITE_API_BASE_URL as string

export const token = {
  get: () => localStorage.getItem("access_token"),
  set: (t: UserTokenPair) => {
    localStorage.setItem("access_token", t.accessToken)
    localStorage.setItem("refresh_token", t.refreshToken)
  },
  getRefresh: () => localStorage.getItem("refresh_token"),
  clear: () => {
    localStorage.removeItem("access_token")
    localStorage.removeItem("refresh_token")
  },
}

export const api = axios.create({ baseURL })

api.interceptors.request.use((req) => {
  const t = token.get()
  if (t) req.headers.Authorization = `Bearer ${t}`
  return req
})

createAuthRefreshInterceptor(api, async (failedRequest) => {
  try {
    const data = await userServiceRefreshToken({
      refreshToken: token.getRefresh() as string,
    })
    token.set(data.tokens)
    failedRequest.response.config.headers.Authorization = `Bearer ${data.tokens.accessToken}`
    return Promise.resolve()
  } catch (e) {
    token.clear()
    return Promise.reject(e)
  }
})
