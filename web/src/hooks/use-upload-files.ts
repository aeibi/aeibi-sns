import { useMutation } from "@tanstack/react-query"
import imageCompression from "browser-image-compression"
import { type FileUploadFileResponse, fileServiceUploadFile } from "@/api/generated"

type UploadFilesVariables = {
  files: File[]
  onFileUploaded?: (payload: { file: File; response: FileUploadFileResponse; index: number }) => void
}

type UploadFilesResult = {
  responses: FileUploadFileResponse[]
  urls: string[]
}

export function useUploadFiles() {
  return useMutation({
    mutationKey: ["uploadFiles"],
    mutationFn: async ({ files, onFileUploaded }: UploadFilesVariables): Promise<UploadFilesResult> => {
      const responses: FileUploadFileResponse[] = []

      for (const [index, file] of files.entries()) {
        const uploadFile = await compressIfImage(file)
        const base64 = await fileToBase64(uploadFile)
        const response = await fileServiceUploadFile({ name: file.name, contentType: file.type, data: base64 })
        responses.push(response)
        onFileUploaded?.({ file, response, index })
      }

      return { responses, urls: responses.map((response) => response.url) }
    },
  })
}

export async function compressIfImage(file: File) {
  if (!file.type.startsWith("image/")) return file
  return imageCompression(file, { maxSizeMB: 0.5 })
}

export function fileToBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()

    reader.onload = () => {
      const result = reader.result as string
      resolve(result.split(",")[1])
    }

    reader.onerror = () => reject(reader.error)
    reader.readAsDataURL(file)
  })
}
