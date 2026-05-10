import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Shield, RefreshCw } from 'lucide-react'
import { api } from '@/lib/api'

interface AuditEntry {
  id: number
  username: string
  action: string
  target: string
  ip: string
  created_at: string
}

const ACTION_COLORS: Record<string, string> = {
  login: 'text-green-400',
  logout: 'text-gray-400',
  totp_enabled: 'text-blue-400',
  totp_disabled: 'text-yellow-400',
  service_start: 'text-green-400',
  service_stop: 'text-red-400',
  service_restart: 'text-yellow-400',
  service_enable: 'text-blue-400',
  service_disable: 'text-orange-400',
  package_install: 'text-blue-400',
  package_remove: 'text-red-400',
  user_create: 'text-green-400',
  user_delete: 'text-red-400',
  firewall_add: 'text-blue-400',
  firewall_delete: 'text-red-400',
  reboot: 'text-red-500',
  terminal_open: 'text-purple-400',
  terminal_close: 'text-gray-400',
  session_revoke: 'text-orange-400',
}

function fmt(ts: string): string {
  try {
    return new Date(ts).toLocaleString()
  } catch {
    return ts
  }
}

export default function AuditLog() {
  const [filter, setFilter] = useState('')

  const { data: entries, isLoading, refetch, isFetching } = useQuery<AuditEntry[]>({
    queryKey: ['audit-log'],
    queryFn: () => api.get<AuditEntry[]>('/api/admin/audit?limit=500'),
    refetchInterval: 30_000,
  })

  const filtered = entries?.filter(e =>
    !filter ||
    e.action.includes(filter) ||
    e.username.includes(filter) ||
    e.target.includes(filter) ||
    e.ip.includes(filter)
  )

  return (
    <div className="p-6 space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold text-white">Audit Log</h1>
          <p className="text-sm text-gray-400 mt-1">All state-changing actions on this server</p>
        </div>
        <button
          onClick={() => refetch()}
          disabled={isFetching}
          className="flex items-center gap-2 px-3 py-1.5 bg-gray-800 hover:bg-gray-700 text-gray-300 text-sm rounded-lg transition-colors disabled:opacity-50"
        >
          <RefreshCw size={14} className={isFetching ? 'animate-spin' : ''} />
          Refresh
        </button>
      </div>

      <input
        type="text"
        value={filter}
        onChange={e => setFilter(e.target.value)}
        placeholder="Filter by action, user, target, or IP…"
        className="w-full bg-gray-900 border border-gray-800 rounded-lg px-3 py-2 text-sm text-white placeholder-gray-600 focus:outline-none focus:border-blue-500"
      />

      <div className="bg-gray-900 border border-gray-800 rounded-lg overflow-hidden">
        <div className="grid grid-cols-[1fr_1fr_1fr_1fr_1fr] gap-0 px-4 py-2 border-b border-gray-800 text-xs text-gray-500 font-medium">
          <span>Time</span>
          <span>User</span>
          <span>Action</span>
          <span>Target</span>
          <span>IP</span>
        </div>

        {isLoading && (
          <div className="p-8 text-center text-gray-500">Loading…</div>
        )}
        {!isLoading && filtered?.length === 0 && (
          <div className="p-8 text-center">
            <Shield size={32} className="text-gray-700 mx-auto mb-3" />
            <p className="text-gray-500 text-sm">No audit entries found</p>
          </div>
        )}

        <div className="divide-y divide-gray-800/40 overflow-y-auto max-h-[600px]">
          {filtered?.map(e => (
            <div
              key={e.id}
              className="grid grid-cols-[1fr_1fr_1fr_1fr_1fr] gap-0 px-4 py-2.5 text-xs hover:bg-gray-800/40 transition-colors"
            >
              <span className="text-gray-500 font-mono">{fmt(e.created_at)}</span>
              <span className="text-gray-300 font-medium">{e.username || '—'}</span>
              <span className={ACTION_COLORS[e.action] ?? 'text-gray-300'}>
                {e.action}
              </span>
              <span className="text-gray-400 font-mono truncate">{e.target || '—'}</span>
              <span className="text-gray-500 font-mono">{e.ip || '—'}</span>
            </div>
          ))}
        </div>
      </div>

      {entries && (
        <p className="text-xs text-gray-600 text-right">
          {filtered?.length} / {entries.length} entries
        </p>
      )}
    </div>
  )
}
