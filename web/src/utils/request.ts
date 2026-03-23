import type { AxiosError, AxiosInstance, AxiosRequestConfig, InternalAxiosRequestConfig } from 'axios'
import axios from 'axios'

interface CustomAxiosInstance extends Omit<AxiosInstance, 'get' | 'post' | 'put' | 'patch' | 'delete'> {
  get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>
  post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  patch<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>
}

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 600000, // 10分钟超时，匹配后端AI生成接口
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json'
  }
}) as CustomAxiosInstance

request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    return config
  },
  (error: AxiosError) => {
    return Promise.reject(error)
  }
)

request.interceptors.response.use(
  (response) => {
    const res = response.data
    if (res.success) {
      return res.data
    } else {
      // 不在这里显示错误提示，让业务代码自行处理
      return Promise.reject(new Error(res.error?.message || '请求失败'))
    }
  },
  async (error: AxiosError<any>) => {
    const status = error.response?.status
    const requestUrl = error.config?.url || ''
    const shouldRedirect =
      status === 401 &&
      !requestUrl.includes('/auth/login') &&
      !window.location.pathname.startsWith('/login')

    if (shouldRedirect) {
      const [{ clearAuthState }, { default: router }] = await Promise.all([
        import('@/stores/auth'),
        import('@/router')
      ])

      clearAuthState()

      if (router.currentRoute.value.path !== '/login') {
        router.push({
          path: '/login',
          query: {
            redirect: router.currentRoute.value.fullPath
          }
        })
      }
    }

    return Promise.reject(error)
  }
)

export default request
