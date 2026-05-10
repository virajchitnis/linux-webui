import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { Play, Square, RotateCcw, Power, PowerOff, Search } from 'lucide-react'

interface Service {
  name: string
  description: string
  load_state: string
  active_state: string
  sub_state: string
}

const activeColors: Record<string, string> = {
  active:       'bg-green-500/20 text-green-400 border-green-500/30',
  activating:   'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
  deactivating: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
  inactive:     'bg-gray-500/20 text-gray-400 border-gray-500/30',
  failed:       'bg-red-500/20 text-red-400 border-red-500/30',
}

function StateBadge({ state }: { state: string }) {
  const cls = activeColors[state] ?? activeColors.inactive
  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${cls}`}>
      {state}
    </span>
  )
}

export default function Services() {
  const [search, setSearch] = useState('')
  const [toast, setToast] = useState<string | null>(null)
  const qc = useQueryClient()

  const { data: services, isLoading, error } = useQuery<Service[]>({
    queryKey: ['services'],
    queryFn: () => api.get('/api/services?type=service'),
    refetchInterval: 10_000,
  })

  const notify = (msg: string) => {
    setToast(msg)
    setTimeout(() => setToast(null), 3000)
  }

  const action = (name: string, act: string) =>
    api.post(`/api/services/${encodeURIComponent(name)}/${act}`, {})
      .then(() => { notify(`${act} sent to ${name}`); qc.invalidateQueries({ queryKey: ['services'] }) })
      .catch((e: Error) => notify(`Error: ${e.message}`))

  const { mutate: start }   = useMutation({ mutationFn: (n: string) => action(n, 'start') })
  const { mutate: stop }    = useMutation({ mutationFn: (n: string) => action(n, 'stop') })
  const { mutate: restart } = useMutation({ mutationFn: (n: string) => action(n, 'restart') })
  const { mutate: enable }  = useMutation({ mutationFn: (n: string) => action(n, 'enable') })
  const { mutate: disable } = useMutation({ mutationFn: (n: string) => action(n, 'disable') })

  const filtered = (services ?? []).filter(s =>
    s.name.toLowerCase().includes(search.toLowerCase()) ||
    s.description.toLowerCase().includes(search.toLowerCase())
  )

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Services</h1>
          <p className="text-gray-400 text-sm mt-0.5">
            {services ? `${services.length} units` : 'Loading…'}
          </p>
        </div>
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-500" />
          <input
            type="text"
            placeholder="Search services…"
            value={search}
            onChange={e => setSearch(e.target.value)}
            className="pl-9 pr-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-sm text-white placeholder-gray-500 focus:outline-none focus:border-blue-500 w-64"
          />
        </div>
      </div>

      {error && (
        <div className="bg-red-900/20 border border-red-800 rounded-lg p-4 text-red-400 text-sm">
          D-Bus unavailable — service management requires systemd.
        </div>
      )}

      <div className="bg-gray-900 rounded-xl border border-gray-800 overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-gray-800 text-xs text-gray-500 uppercase tracking-wide">
              <th className="text-left px-4 py-3 font-medium">Name</th>
              <th className="text-left px-4 py-3 font-medium">Description</th>
              <th className="text-left px-4 py-3 font-medium">State</th>
              <th className="text-left px-4 py-3 font-medium">Sub-state</th>
              <th className="text-right px-4 py-3 font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            {isLoading && (
              <tr><td colSpan={5} className="px-4 py-8 text-center text-gray-500">Loading services…</td></tr>
            )}
            {!isLoading && filtered.length === 0 && (
              <tr><td colSpan={5} className="px-4 py-8 text-center text-gray-500">No services found.</td></tr>
            )}
            {filtered.map(svc => (
              <tr key={svc.name} className="border-b border-gray-800/50 hover:bg-gray-800/30 transition-colors">
                <td className="px-4 py-3 font-mono text-xs text-blue-300">{svc.name}</td>
                <td className="px-4 py-3 text-gray-400 max-w-xs truncate">{svc.description}</td>
                <td className="px-4 py-3"><StateBadge state={svc.active_state} /></td>
                <td className="px-4 py-3 text-gray-500 text-xs">{svc.sub_state}</td>
                <td className="px-4 py-3">
                  <div className="flex items-center justify-end gap-1">
                    <ActionBtn title="Start" onClick={() => start(svc.name)} disabled={svc.active_state === 'active'}>
                      <Play className="h-3.5 w-3.5" />
                    </ActionBtn>
                    <ActionBtn title="Stop" onClick={() => stop(svc.name)} disabled={svc.active_state === 'inactive'}>
                      <Square className="h-3.5 w-3.5" />
                    </ActionBtn>
                    <ActionBtn title="Restart" onClick={() => restart(svc.name)}>
                      <RotateCcw className="h-3.5 w-3.5" />
                    </ActionBtn>
                    <ActionBtn title="Enable" onClick={() => enable(svc.name)}>
                      <Power className="h-3.5 w-3.5" />
                    </ActionBtn>
                    <ActionBtn title="Disable" onClick={() => disable(svc.name)}>
                      <PowerOff className="h-3.5 w-3.5" />
                    </ActionBtn>
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

function ActionBtn({
  title, onClick, disabled, children,
}: {
  title: string
  onClick: () => void
  disabled?: boolean
  children: React.ReactNode
}) {
  return (
    <button
      title={title}
      onClick={onClick}
      disabled={disabled}
      className="p-1.5 rounded hover:bg-gray-700 text-gray-400 hover:text-white disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
    >
      {children}
    </button>
  )
}
