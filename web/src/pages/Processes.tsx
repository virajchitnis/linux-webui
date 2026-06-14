import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { Search, Square, Gauge } from 'lucide-react'

interface Process {
  pid: number
  ppid: number
  name: string
  state: string
  user: string
  mem_rss_kb: number
  cpu_pct: number
  command: string
  nice: number
}

const STATE_COLORS: Record<string, string> = {
  R: 'text-green-400',
  S: 'text-blue-400',
  D: 'text-yellow-400',
  Z: 'text-red-400',
  T: 'text-gray-500',
  I: 'text-gray-500',
}

function fmtMem(kb: number): string {
  if (kb >= 1024 * 1024) return `${(kb / 1024 / 1024).toFixed(1)}G`
  if (kb >= 1024) return `${(kb / 1024).toFixed(0)}M`
  return `${kb}K`
}

export default function Processes() {
  const [search, setSearch] = useState('')
  const [showAll, setShowAll] = useState(false)
  const [toastMsg, setToastMsg] = useState<string | null>(null)
  const qc = useQueryClient()

  const { data: procs, isLoading } = useQuery<Process[]>({
    queryKey: ['processes'],
    queryFn: () => api.get('/api/processes'),
    refetchInterval: 5000,
  })

  const toast = (msg: string) => {
    setToastMsg(msg)
    setTimeout(() => setToastMsg(null), 3000)
  }

  const killProc = async (pid: number, signal = 'SIGTERM') => {
    if (!window.confirm(`Send ${signal} to PID ${pid}?`)) return
    try {
      await api.post(`/api/processes/${pid}/kill`, { signal })
      toast(`${signal} sent to PID ${pid}`)
      qc.invalidateQueries({ queryKey: ['processes'] })
    } catch (e) {
      toast(`Error: ${(e as Error).message}`)
    }
  }

  const filtered = (procs ?? [])
    .filter(p => showAll || (!p.command.startsWith('[') || p.pid < 10))
    .filter(p =>
      !search ||
      p.name.toLowerCase().includes(search.toLowerCase()) ||
      p.user.toLowerCase().includes(search.toLowerCase()) ||
      p.command.toLowerCase().includes(search.toLowerCase()) ||
      String(p.pid).includes(search)
    )
    .sort((a, b) => b.mem_rss_kb - a.mem_rss_kb)

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Processes</h1>
          <p className="text-gray-400 text-sm mt-0.5">
            {procs ? `${procs.length} total` : 'Loading…'}
            {filtered.length !== procs?.length ? `, ${filtered.length} shown` : ''}
          </p>
        </div>
        <div className="flex items-center gap-3">
          <label className="flex items-center gap-2 text-xs text-gray-400 cursor-pointer">
            <input
              type="checkbox"
              checked={showAll}
              onChange={e => setShowAll(e.target.checked)}
              className="rounded"
            />
            Show kernel threads
          </label>
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-500" />
            <input
              type="text"
              placeholder="Filter by name, user, PID…"
              value={search}
              onChange={e => setSearch(e.target.value)}
              className="pl-9 pr-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-sm text-white placeholder-gray-500 focus:outline-none focus:border-blue-500 w-64"
            />
          </div>
        </div>
      </div>

      <div className="bg-gray-900 rounded-xl border border-gray-800 overflow-hidden">
        <table className="w-full text-xs">
          <thead>
            <tr className="text-gray-500 uppercase tracking-wide border-b border-gray-800">
              <th className="text-right px-3 py-2 font-medium w-16">PID</th>
              <th className="text-left px-3 py-2 font-medium w-20">User</th>
              <th className="text-left px-3 py-2 font-medium w-6">S</th>
              <th className="text-right px-3 py-2 font-medium w-16">CPU%</th>
              <th className="text-right px-3 py-2 font-medium w-16">MEM</th>
              <th className="text-left px-3 py-2 font-medium">Command</th>
              <th className="text-right px-3 py-2 font-medium w-20">Actions</th>
            </tr>
          </thead>
          <tbody>
            {isLoading && (
              <tr><td colSpan={7} className="px-4 py-8 text-center text-gray-500">Loading processes…</td></tr>
            )}
            {filtered.map(p => (
              <tr key={p.pid} className="border-b border-gray-800/50 hover:bg-gray-800/20 transition-colors">
                <td className="px-3 py-1.5 text-right text-gray-500 font-mono">{p.pid}</td>
                <td className="px-3 py-1.5 text-purple-400 truncate max-w-[80px]">{p.user}</td>
                <td className="px-3 py-1.5 font-mono font-bold">
                  <span className={STATE_COLORS[p.state] ?? 'text-gray-400'}>{p.state}</span>
                </td>
                <td className="px-3 py-1.5 text-right font-mono text-gray-300">
                  {p.cpu_pct > 0 ? p.cpu_pct.toFixed(1) : '–'}
                </td>
                <td className="px-3 py-1.5 text-right font-mono text-gray-300">
                  {p.mem_rss_kb > 0 ? fmtMem(p.mem_rss_kb) : '–'}
                </td>
                <td className="px-3 py-1.5 text-gray-400 max-w-xs truncate font-mono" title={p.command}>
                  {p.command}
                </td>
                <td className="px-3 py-1.5">
                  <div className="flex items-center justify-end gap-1">
                    <button
                      title="Send SIGTERM"
                      onClick={() => killProc(p.pid, 'SIGTERM')}
                      disabled={p.pid <= 1}
                      className="p-1 rounded hover:bg-yellow-900/40 text-yellow-500 hover:text-yellow-400 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                    >
                      <Square className="h-3 w-3" />
                    </button>
                    <button
                      title="Send SIGKILL"
                      onClick={() => killProc(p.pid, 'SIGKILL')}
                      disabled={p.pid <= 1}
                      className="p-1 rounded hover:bg-red-900/40 text-red-500 hover:text-red-400 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
                    >
                      <Gauge className="h-3 w-3" />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {toastMsg && (
        <div className="fixed bottom-4 right-4 bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-sm text-white shadow-xl">
          {toastMsg}
        </div>
      )}
    </div>
  )
}
