import request from '../utils/request'

export interface ImageStyle {
  id: number
  name_zh: string
  name_en: string
  style_value: string
  sort_order: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface CreateStyleRequest {
  name_zh: string
  name_en: string
  style_value: string
  sort_order?: number
  is_active?: boolean
}

export interface UpdateStyleRequest {
  name_zh?: string
  name_en?: string
  style_value?: string
  sort_order?: number
  is_active?: boolean
}

export const styleAPI = {
  list(all = false) {
    return request.get<ImageStyle[]>('/styles', { params: all ? { all: 'true' } : {} })
  },

  create(data: CreateStyleRequest) {
    return request.post<ImageStyle>('/styles', data)
  },

  update(id: number, data: UpdateStyleRequest) {
    return request.put<ImageStyle>(`/styles/${id}`, data)
  },

  delete(id: number) {
    return request.delete(`/styles/${id}`)
  }
}
