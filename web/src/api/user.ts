import request from '@/utils/request'

export type UserRole = 'admin' | 'user'

export interface UserItem {
  id: number
  username: string
  role: UserRole | string
}

export interface CreateUserPayload {
  username: string
  password: string
  role: UserRole
}

export interface UpdateUserPayload {
  username?: string
  password?: string
  role?: UserRole
}

export const userAPI = {
  list() {
    return request.get<UserItem[]>('/users')
  },

  create(data: CreateUserPayload) {
    return request.post<UserItem>('/users', data)
  },

  update(id: number, data: UpdateUserPayload) {
    return request.put<UserItem>(`/users/${id}`, data)
  },

  delete(id: number) {
    return request.delete<void>(`/users/${id}`)
  }
}
