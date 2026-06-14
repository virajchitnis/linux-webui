import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { UserPlus, Trash2, Key } from 'lucide-react'

interface SysUser {
  username: string
  uid: number
  gid: number
  gecos: string
  home: string
  shell: string
  system: boolean
}

export default function Users() {
  const [showSystem, setShowSystem] = useState(false)
  const [toast, setToast] = useState<string | null>(null)
  const [createOpen, setCreateOpen] = useState(false)
  const [newUser, setNewUser] = useState({ username: '', password: '' })
  const [pwdTarget, setPwdTarget] = useState<string | null>(null)
  const [newPwd, setNewPwd] = useState('')
  const qc = useQueryClient()

  const { data: userList, isLoading } = useQuery<SysUser[]>({
    queryKey: ['sys-users'],
    queryFn: () => api.get('/api/users'),
    refetchInterval: 15_000,
  })

  const notify = (msg: string) => { setToast(msg); setTimeout(() => setToast(null), 3500) }

  const { mutate: createUser, isPending: creating } = useMutation({
    mutationFn: () => api.post('/api/users', { username: newUser.username, password: newUser.password }),
    onSuccess: () => {
      notify(`User ${newUser.username} created`)
      setCreateOpen(false)
      setNewUser({ username: '', password: '' })
      qc.invalidateQueries({ queryKey: ['sys-users'] })
    },
    onError: (e: Error) => notify(`Error: ${e.message}`),
  })

  const { mutate: deleteUser } = useMutation({
    mutationFn: (username: string) => api.delete(`/api/users/${encodeURIComponent(username)}`),
    onSuccess: (_, username) => {
      notify(`User ${username} deleted`)
      qc.invalidateQueries({ queryKey: ['sys-users'] })
    },
    onError: (e: Error) => notify(`Error: ${e.message}`),
  })

  const { mutate: setPassword, isPending: settingPwd } = useMutation({
    mutationFn: ({ username, pwd }: { username: string; pwd: string }) =>
      api.post(`/api/users/${encodeURIComponent(username)}/password`, { password: pwd }),
    onSuccess: (_, { username }) => {
      notify(`Password updated for ${username}`)
      setPwdTarget(null)
      setNewPwd('')
    },
    onError: (e: Error) => notify(`Error: ${e.message}`),
  })

  const confirmDelete = (username: string) => {
    if (window.confirm(`Delete user ${username}? This will remove their home directory.`)) {
      deleteUser(username)
    }
  }

  const visible = (userList ?? []).filter(u => showSystem || !u.system)

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Users</h1>
          <p className="text-gray-400 text-sm mt-0.5">System user accounts</p>
        </div>
        <div className="flex items-center gap-3">
          <label className="flex items-center gap-2 text-xs text-gray-400 cursor-pointer">
            <input type="checkbox" checked={showSystem} onChange={e => setShowSystem(e.target.checked)} />
            Show system users
          </label>
          <button
            onClick={() => setCreateOpen(true)}
            className="flex items-center gap-2 px-3 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm rounded-lg transition-colors"
          >
            <UserPlus className="h-4 w-4" />
            Create user
          </button>
        </div>
      </div>

      {/* Create user dialog */}
      {createOpen && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-gray-900 border border-gray-700 rounded-xl p-6 w-80 space-y-4">
            <h2 className="text-white font-semibold">Create user</h2>
            <div className="space-y-3">
              <input
                autoFocus
                placeholder="Username"
                value={newUser.username}
                onChange={e => setNewUser(u => ({ ...u, username: e.target.value }))}
                className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white text-sm focus:outline-none focus:border-blue-500"
              />
              <input
                type="password"
                placeholder="Initial password (optional)"
                value={newUser.password}
                onChange={e => setNewUser(u => ({ ...u, password: e.target.value }))}
                className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white text-sm focus:outline-none focus:border-blue-500"
              />
            </div>
            <div className="flex justify-end gap-2">
              <button onClick={() => setCreateOpen(false)} className="px-3 py-2 text-sm text-gray-400 hover:text-white">Cancel</button>
              <button
                onClick={() => createUser()}
                disabled={creating || !newUser.username}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm rounded-lg disabled:opacity-50"
              >
                {creating ? 'Creating…' : 'Create'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Set password dialog */}
      {pwdTarget && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-gray-900 border border-gray-700 rounded-xl p-6 w-80 space-y-4">
            <h2 className="text-white font-semibold">Set password for <span className="text-blue-400">{pwdTarget}</span></h2>
            <input
              autoFocus
              type="password"
              placeholder="New password (min 8 chars)"
              value={newPwd}
              onChange={e => setNewPwd(e.target.value)}
              className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white text-sm focus:outline-none focus:border-blue-500"
            />
            <div className="flex justify-end gap-2">
              <button onClick={() => { setPwdTarget(null); setNewPwd('') }} className="px-3 py-2 text-sm text-gray-400 hover:text-white">Cancel</button>
              <button
                onClick={() => setPassword({ username: pwdTarget, pwd: newPwd })}
                disabled={settingPwd || newPwd.length < 8}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm rounded-lg disabled:opacity-50"
              >
                {settingPwd ? 'Setting…' : 'Set password'}
              </button>
            </div>
          </div>
        </div>
      )}

      <div className="bg-gray-900 rounded-xl border border-gray-800 overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="text-xs text-gray-500 uppercase tracking-wide border-b border-gray-800">
              <th className="text-left px-4 py-2 font-medium">Username</th>
              <th className="text-left px-4 py-2 font-medium w-16">UID</th>
              <th className="text-left px-4 py-2 font-medium">Home</th>
              <th className="text-left px-4 py-2 font-medium">Shell</th>
              <th className="text-left px-4 py-2 font-medium">GECOS</th>
              <th className="text-right px-4 py-2 font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            {isLoading && <tr><td colSpan={6} className="px-4 py-8 text-center text-gray-500">Loading users…</td></tr>}
            {visible.map(u => (
              <tr key={u.username} className="border-b border-gray-800/50 hover:bg-gray-800/20">
                <td className="px-4 py-2 font-mono text-xs text-blue-300">{u.username}</td>
                <td className="px-4 py-2 text-gray-500 text-xs">{u.uid}</td>
                <td className="px-4 py-2 text-gray-400 text-xs font-mono">{u.home}</td>
                <td className="px-4 py-2 text-gray-500 text-xs font-mono">{u.shell}</td>
                <td className="px-4 py-2 text-gray-500 text-xs">{u.gecos}</td>
                <td className="px-4 py-2">
                  <div className="flex items-center justify-end gap-1">
                    <button title="Set password" onClick={() => { setPwdTarget(u.username); setNewPwd('') }}
                      className="p-1.5 rounded hover:bg-gray-700 text-gray-400 hover:text-white transition-colors">
                      <Key className="h-3.5 w-3.5" />
                    </button>
                    <button title="Delete user" onClick={() => confirmDelete(u.username)}
                      className="p-1.5 rounded hover:bg-red-900/40 text-gray-400 hover:text-red-400 transition-colors">
                      <Trash2 className="h-3.5 w-3.5" />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {toast && (
        <div className="fixed bottom-4 right-4 bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-sm text-white shadow-xl">
          {toast}
        </div>
      )}
    </div>
  )
}
