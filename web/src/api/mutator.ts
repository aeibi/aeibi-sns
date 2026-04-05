import type { AxiosError, AxiosRequestConfig } from "axios"
import { api } from "./client"

export const customInstance = async <T>(config: AxiosRequestConfig, options?: AxiosRequestConfig): Promise<T> => {
  const res = await api({ ...config, ...options })
  return res.data as T
}

export type ErrorType<E> = AxiosError<E>
