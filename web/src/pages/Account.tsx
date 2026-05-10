import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useAuth } from '@/hooks/useAuth'
import { Key, Trash2, Monitor } from 'lucide-react'

interface SessionInfo {
  id: string
  ip: string
  user_agent: string
  created_at: string
  last_seen: string
}

export default function Account() {
  const { data: me } = useAuth()
  const qc = useQueryClient()
  const [toast, setToast] = useState<string | null>(null)

  // Password change
  const [pwd, setPwd] = useState({ current: '', next: '', confirm: '' })
  const notify = (msg: string) => { setToast(msg); setTimeout(() => setToast(null), 3500) }

  const { mutate: changePwd, isPending: changingPwd } = useMutation({
    mutationFn: () => api.post('/api/account/password', { current_password: pwd.current, new_password: pwd.next }),
    onSuccess: () => { notify('Password changed'); setPwd({ current: '', next: '', confirm: '' }) },
    onError: (e: Error) => notify(`Error: ${e.message}`),
  })

  const pwdValid = pwd.next.length >= 8 && pwd.next === pwd.confirm && pwd.current.length > 0

  // Own sessions
  const { data: sessions } = useQuery<SessionInfo[]>({
    queryKey: ['own-sessions'],
    queryFn: () => api.get('/api/account/sessions'),
  })

  const { mutate: revokeSession } = useMutation({
    mutationFn: (id: string) => api.delete(`/api/account/sessions/${id}`),
    onSuccess: () => { notify('Session revoked'); qc.invalidateQueries({ queryKey: ['own-sessions'] }) },
    onError: (e: Error) => notify(`Error: ${e.message}`),
  })

  return (
    <div className="space-y-6 max-w-xl">
      <div>
        <h1 className="text-2xl font-bold text-white">Account</h1>
        <p className="text-gray-400 text-sm mt-0.5">{me?.username} · {me?.role}</p>
      </div>

      {/* Change password */}
      <section className="bg-gray-900 rounded-xl border border-gray-800 p-5 space-y-4">
        <h2 className="text-sm font-semibold text-gray-300 flex items-center gap-2">
          <Key className="h-4 w-4" />
          Change password
        </h2>
        <div className="space-y-3">
          <input
            type="password"
            placeholder="Current password"
            value={pwd.current}
            onChange={e => setPwd(p => ({ ...p, current: e.target.value }))}
            className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white text-sm focus:outline-none focus:border-blue-500"
          />
          <input
            type="password"
            placeholder="New password (min 8 chars)"
            value={pwd.next}
            onChange={e => setPwd(p => ({ ...p, next: e.target.value }))}
            className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white text-sm focus:outline-none focus:border-blue-500"
          />
          <input
            type="password"
            placeholder="Confirm new password"
            value={pwd.confirm}
            onChange={e => setPwd(p => ({ ...p, confirm: e.target.value }))}
            className={`w-full px-3 py-2 bg-gray-800 border rounded-lg text-white text-sm focus:outline-none ${
              pwd.confirm && pwd.confirm !== pwd.next ? 'border-red-500' : 'border-gray-700 focus:border-blue-500'
            }`}
          />
        </div>
        {pwd.confirm && pwd.confirm !== pwd.next && (
          <p className="text-red-400 text-xs">Passwords do not match</p>
        )}
        <button
          onClick={() => changePwd()}
          disabled={changingPwd || !pwdValid}
          className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm rounded-lg disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {changingPwd ? 'Changing…' : 'Change password'}
        </button>
      </section>

      {/* Active sessions */}
      <section className="bg-gray-900 rounded-xl border border-gray-800 overflow-hidden">
        <div className="px-5 py-4 border-b border-gray-800">
          <h2 className="text-sm font-semibold text-gray-300 flex items-center gap-2">
            <Monitor className="h-4 w-4" />
            Active sessions
          </h2>
        </div>
        {!sessions?.length && (
          <div className="px-5 py-4 text-gray-500 text-sm">No sessions found.</div>
        )}
        {sessions?.map(s => (
          <div key={s.id} className="flex items-start justify-between px-5 py-3 border-b border-gray-800/50 last:border-0">
            <div className="space-y-0.5 text-xs">
              <p className="text-gray-300 font-mono">{s.ip}</p>
              <p className="text-gray-500 truncate max-w-xs">{s.user_agent}</p>
              <p className="text-gray-600">Last seen: {new Date(s.last_seen).toLocaleString()}</p>
            </div>
            <button
              onClick={() => revokeSession(s.id)}
              title="Revoke session"
              className="p-1.5 rounded hover:bg-red-900/40 text-gray-500 hover:text-red-400 transition-colors shrink-0"
            >
              <Trash2 className="h-3.5 w-3.5" />
            </button>
          </div>
        ))}
      </section>

      {toast && (
        <div className="fixed bottom-4 right-4 bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-sm text-white shadow-xl">
          {toast}
        </div>
      )}
    </div>
  )
}
