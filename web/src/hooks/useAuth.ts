import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'

interface Me {
  username: string
  role: 'admin' | 'readonly'
}

export function useAuth() {
  return useQuery<Me | null>({
    queryKey: ['me'],
    queryFn: async () => {
      try {
        return await api.get<Me>('/api/auth/me')
      } catch {
        return null
      }
    },
    staleTime: 30_000,
  })
}

export function useLogin() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (creds: { username: string; password: string }) =>
      api.post<Me>('/api/auth/login', creds),
    onSuccess: (data) => {
      qc.setQueryData(['me'], data)
    },
  })
}

export function useLogout() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => api.post('/api/auth/logout'),
    onSuccess: () => {
      qc.setQueryData(['me'], null)
      qc.clear()
    },
  })
}
