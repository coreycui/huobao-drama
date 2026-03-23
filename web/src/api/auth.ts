import request from '@/utils/request'

export interface AuthUser {
  id: number
  username: string
  role: string
}

export interface LoginPayload {
  username: string
  password: string
}

export interface ChangePasswordPayload {
  old_password: string
  new_password: string
}

export const authAPI = {
  login(data: LoginPayload) {
    return request.post<void>('/auth/login', data, {
      withCredentials: true
    })
  },

  logout() {
    return request.post<void>('/auth/logout')
  },

  me() {
    return request.get<AuthUser>('/auth/me')
  },

  changePassword(data: ChangePasswordPayload) {
    return request.put<void>('/auth/password', data)
  }
}
