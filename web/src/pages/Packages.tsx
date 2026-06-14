import { useState, useEffect, useRef, useCallback } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { RefreshCw, ArrowUpCircle, Package } from 'lucide-react'

interface PkgInfo {
  name: string
  available: string
  current: string
  arch: string
}

interface AptStatus {
  running: boolean
}

type OutputLine = { text: string; id: number }

let lineIdSeq = 0

function useAptWS() {
  const [lines, setLines] = useState<OutputLine[]>([])
  const [busy, setBusy] = useState(false)
  const [done, setDone] = useState<{ exit_code: number; error: string } | null>(null)
  const wsRef = useRef<WebSocket | null>(null)

  const run = useCallback((action: string, pkg?: string) => {
    if (wsRef.current) return
    setLines([])
    setDone(null)
    setBusy(true)

    const proto = window.location.protocol === 'https:' ? 'wss' : 'ws'
    const ws = new WebSocket(`${proto}://${window.location.host}/ws/apt`)
    wsRef.current = ws

    ws.onopen = () => {
      ws.send(JSON.stringify({ type: 'run', payload: { action, package: pkg ?? '' } }))
    }
    ws.onmessage = (ev) => {
      const msg = JSON.parse(ev.data as string)
      if (msg.type === 'output') {
        setLines(prev => [...prev, { text: msg.payload.line as string, id: lineIdSeq++ }])
      } else if (msg.type === 'done') {
        setDone(msg.payload as { exit_code: number; error: string })
        setBusy(false)
        wsRef.current = null
      } else if (msg.type === 'error') {
        setLines(prev => [...prev, { text: `ERROR: ${(msg.payload as { message: string }).message}`, id: lineIdSeq++ }])
        setBusy(false)
        wsRef.current = null
      }
    }
    ws.onerror = () => {
      setLines(prev => [...prev, { text: 'WebSocket connection error', id: lineIdSeq++ }])
      setBusy(false)
      wsRef.current = null
    }
    ws.onclose = () => { wsRef.current = null }
  }, [])

  return { lines, busy, done, run }
}

export default function Packages() {
  const qc = useQueryClient()
  const outputRef = useRef<HTMLDivElement>(null)
  const { lines, busy, done, run } = useAptWS()

  const { data: pkgs, isLoading, refetch } = useQuery<PkgInfo[]>({
    queryKey: ['upgradable'],
    queryFn: () => api.get('/api/packages/upgradable'),
  })

  const { data: aptStatus } = useQuery<AptStatus>({
    queryKey: ['apt-status'],
    queryFn: () => api.get('/api/packages/status'),
    refetchInterval: 2000,
  })

  useEffect(() => {
    if (outputRef.current) {
      outputRef.current.scrollTop = outputRef.current.scrollHeight
    }
  }, [lines])

  const doUpdate = () => {
    run('update')
    setTimeout(() => { refetch(); qc.invalidateQueries({ queryKey: ['upgradable'] }) }, 6000)
  }

  const doUpgrade = () => run('upgrade')
  const isBusy = busy || aptStatus?.running

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Packages</h1>
          <p className="text-gray-400 text-sm mt-0.5">APT package management</p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={doUpdate}
            disabled={isBusy}
            className="flex items-center gap-2 px-3 py-2 bg-gray-800 hover:bg-gray-700 border border-gray-700 text-white text-sm rounded-lg disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            <RefreshCw className={`h-4 w-4 ${busy ? 'animate-spin' : ''}`} />
            Refresh package lists
          </button>
          <button
            onClick={doUpgrade}
            disabled={isBusy || !pkgs?.length}
            className="flex items-center gap-2 px-3 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm rounded-lg disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            <ArrowUpCircle className="h-4 w-4" />
            Upgrade all ({pkgs?.length ?? 0})
          </button>
        </div>
      </div>

      {/* Streaming output panel */}
      {(lines.length > 0 || busy) && (
        <div className="bg-gray-950 rounded-xl border border-gray-800 overflow-hidden">
          <div className="px-4 py-2 border-b border-gray-800 flex items-center justify-between">
            <span className="text-xs text-gray-500 font-mono">apt-get output</span>
            {done && (
              <span className={`text-xs font-medium ${done.exit_code === 0 ? 'text-green-400' : 'text-red-400'}`}>
                {done.exit_code === 0 ? '✓ Done' : `✗ Failed${done.error ? ': ' + done.error : ''}`}
              </span>
            )}
            {busy && <span className="text-xs text-yellow-400 animate-pulse">Running…</span>}
          </div>
          <div
            ref={outputRef}
            className="font-mono text-xs text-gray-300 p-4 h-64 overflow-y-auto whitespace-pre-wrap leading-relaxed"
          >
            {lines.map(l => <div key={l.id}>{l.text}</div>)}
          </div>
        </div>
      )}

      {/* Upgradable packages table */}
      <div className="bg-gray-900 rounded-xl border border-gray-800 overflow-hidden">
        <div className="px-4 py-3 border-b border-gray-800">
          <h2 className="text-sm font-semibold text-gray-300">
            Upgradable packages
            {pkgs && <span className="ml-2 text-gray-500 font-normal">({pkgs.length})</span>}
          </h2>
        </div>
        {isLoading && (
          <div className="px-4 py-8 text-center text-gray-500 text-sm">Checking for updates…</div>
        )}
        {!isLoading && !pkgs?.length && (
          <div className="px-4 py-8 text-center flex flex-col items-center gap-2 text-gray-500">
            <Package className="h-8 w-8 opacity-30" />
            <span className="text-sm">Everything is up to date</span>
          </div>
        )}
        {pkgs && pkgs.length > 0 && (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-xs text-gray-500 uppercase tracking-wide border-b border-gray-800">
                <th className="text-left px-4 py-2 font-medium">Package</th>
                <th className="text-left px-4 py-2 font-medium">Installed</th>
                <th className="text-left px-4 py-2 font-medium">Available</th>
                <th className="text-left px-4 py-2 font-medium">Arch</th>
              </tr>
            </thead>
            <tbody>
              {pkgs.map(p => (
                <tr key={p.name} className="border-b border-gray-800/50 hover:bg-gray-800/20">
                  <td className="px-4 py-2 font-mono text-xs text-blue-300">{p.name}</td>
                  <td className="px-4 py-2 text-gray-500 font-mono text-xs">{p.current}</td>
                  <td className="px-4 py-2 text-green-400 font-mono text-xs">{p.available}</td>
                  <td className="px-4 py-2 text-gray-500 text-xs">{p.arch}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}
