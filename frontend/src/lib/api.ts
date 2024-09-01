interface EmailVerifyRequest {
  email: string
  code: string
}

interface EmailSendRequest {
  email: string
}

export interface ConfigResponse {
  title: string
  brand: string
  support: string
}

const credentials = "same-origin"

export async function getConfig(): Promise<ConfigResponse> {
  const response = await fetch(`/api/config`, {
    method: "get",
    credentials: credentials,
  })

  return response.json()
}

export async function verifyToken(): Promise<boolean> {
  const response = await fetch(`/api/verify-token`, {
    method: "post",
    credentials: credentials,
  })

  return response.ok
}

export async function emailVerify(email: string, code: string): Promise<boolean> {
  const payload: EmailVerifyRequest = {email, code}
  const response = await fetch(`/api/email/verify`, {
    method: "post",
    credentials: credentials,
    body: JSON.stringify(payload),
  })

  return response.ok
}

export async function emailSend(email: string) {
  const payload: EmailSendRequest = {email}
  await fetch(`/api/email/send`, {
    method: "post",
    credentials: credentials,
    body: JSON.stringify(payload),
  })
}
