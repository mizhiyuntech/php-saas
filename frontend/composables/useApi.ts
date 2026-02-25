export function useApi() {
  const config = useRuntimeConfig()
  const authStore = useAuthStore()

  const baseURL = config.public.apiBase as string || ''

  async function request<T = any>(url: string, options: any = {}): Promise<T> {
    const headers: Record<string, string> = {
      ...(options.headers || {})
    }

    if (authStore.token) {
      headers['Authorization'] = `Bearer ${authStore.token}`
    }

    if (options.body && !(options.body instanceof FormData)) {
      headers['Content-Type'] = 'application/json'
      options.body = JSON.stringify(options.body)
    }

    const response = await fetch(`${baseURL}${url}`, {
      ...options,
      headers
    })

    const data = await response.json()

    if (data.code === 401) {
      authStore.logout()
      navigateTo('/login')
      throw new Error(data.message || '认证已过期')
    }

    return data
  }

  function get<T = any>(url: string, params?: Record<string, any>) {
    let queryString = ''
    if (params) {
      const searchParams = new URLSearchParams()
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '') {
          searchParams.append(key, String(value))
        }
      })
      queryString = searchParams.toString()
      if (queryString) queryString = '?' + queryString
    }
    return request<T>(url + queryString)
  }

  function post<T = any>(url: string, body?: any) {
    return request<T>(url, { method: 'POST', body })
  }

  function put<T = any>(url: string, body?: any) {
    return request<T>(url, { method: 'PUT', body })
  }

  function del<T = any>(url: string) {
    return request<T>(url, { method: 'DELETE' })
  }

  function upload<T = any>(url: string, formData: FormData) {
    return request<T>(url, { method: 'POST', body: formData })
  }

  return { get, post, put, del, upload }
}
