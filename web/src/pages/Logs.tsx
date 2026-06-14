import { useState, useEffect, useRef, useCallback } from 'react'
import { Search, Filter, Wifi, WifiOff } from 'lucide-react'

interface LogEntry {
  timestamp: string
  unit: string
  priority: number
  message: string
  pid: string
}

const PRIORITY_LABELS: Record<number, { label: string; cls: string }> = {
  0: { label: 'EMERG',   cls: 'text-red-400 font-bold' },
  1: { label: 'ALERT',   cls: 'text-red-400 font-bold' },
  2: { label: 'CRIT',    cls: 'text-red-400' },
  3: { label: 'ERR',     cls: 'text-orange-400' },
  4: { label: 'WARN',    cls: 'text-yellow-400' },
  5: { label: 'NOTICE',  cls: 'text-blue-400' },
  6: { label: 'INFO',    cls: 'text-gray-400' },
  7: { label: 'DEBUG',   cls: 'text-gray-600' },
}

const MAX_ENTRIES = 2000

function fmtTime(ts: string): string {
  if (!ts) return ''
  try {
    return new Date(ts).toLocaleTimeString()
  } catch {
    return ts
  }
}

export default function Logs() {
  const [entries, setEntries] = useState<LogEntry[]>([])
  const [connected, setConnected] = useState(false)
  const [ready, setReady] = useState(false)
  const [unit, setUnit] = useState('')
  const [priority, setPriority] = useState('')
  const [search, setSearch] = useState('')
  const [autoScroll, setAutoScroll] = useState(true)
  const bottomRef = useRef<HTMLDivElement>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const unitInput = useRef('')
  const prioInput = useRef('')

  const connect = useCallback((u: string, p: string) => {
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
    setEntries([])
    setReady(false)
    setConnected(false)

    const proto = window.location.protocol === 'https:' ? 'wss' : 'ws'
    const params = new URLSearchParams()
    if (u) params.set('unit', u)
    if (p) params.set('priority', p)
    const url = `${proto}://${window.location.host}/ws/logs${params.size ? '?' + params : ''}`
    const ws = new WebSocket(url)
    wsRef.current = ws

    ws.onopen = () => setConnected(true)
    ws.onmessage = (ev) => {
      const msg = JSON.parse(ev.data as string)
      if (msg.type === 'log') {
        setEntries(prev => {
          const next = [...prev, msg.payload as LogEntry]
          return next.length > MAX_ENTRIES ? next.slice(next.length - MAX_ENTRIES) : next
        })
      } else if (msg.type === 'ready') {
        setReady(true)
      }
    }
    ws.onerror = () => setConnected(false)
    ws.onclose = () => { setConnected(false); wsRef.current = null }
  }, [])

  useEffect(() => {
    connect('', '')
    return () => { wsRef.current?.close() }
  }, [connect])

  useEffect(() => {
    if (autoScroll && bottomRef.current) {
      bottomRef.current.scrollIntoView({ behavior: 'smooth' })
    }
  }, [entries, autoScroll])

  const applyFilter = () => {
    unitInput.current = unit
    prioInput.current = priority
    connect(unit, priority)
  }

  const filtered = search
    ? entries.filter(e =>
        e.message.toLowerCase().includes(search.toLowerCase()) ||
        e.unit.toLowerCase().includes(search.toLowerCase())
      )
    : entries

  return (
    <div className="h-full flex flex-col gap-4">
      <div className="flex items-center justify-between flex-shrink-0">
        <div>
          <h1 className="text-2xl font-bold text-white">Logs</h1>
          <p className="text-gray-400 text-sm mt-0.5">System journal (journalctl)</p>
        </div>
        <div className="flex items-center gap-2">
          {connected
            ? <Wifi className="h-4 w-4 text-green-400" />
            : <WifiOff className="h-4 w-4 text-gray-500" />}
          <span className="text-xs text-gray-500">{connected ? (ready ? 'Live' : 'Loading…') : 'Disconnected'}</span>
        </div>
      </div>

      {/* Filters */}
      <div className="flex gap-2 flex-shrink-0">
        <input
          type="text"
          placeholder="Unit (e.g. nginx.service)"
          value={unit}
          onChange={e => setUnit(e.target.value)}
          onKeyDown={e => e.key === 'Enter' && applyFilter()}
          className="px-3 py-2 bg-gray-900 border border-gray-700 rounded-lg text-sm text-white placeholder-gray-500 focus:outline-none focus:border-blue-500 flex-1 max-w-xs"
        />
        <select
          value={priority}
          onChange={e => setPriority(e.target.value)}
          className="px-3 py-2 bg-gray-900 border border-gray-700 rounded-lg text-sm text-white focus:outline-none focus:border-blue-500"
        >
          <option value="">All priorities</option>
          <option value="3">Error and above</option>
          <option value="4">Warning and above</option>
          <option value="6">Info and above</option>
          <option value="7">Debug (all)</option>
        </select>
        <button
          onClick={applyFilter}
          className="flex items-center gap-1.5 px-3 py-2 bg-gray-800 hover:bg-gray-700 border border-gray-700 text-white text-sm rounded-lg transition-colors"
        >
          <Filter className="h-4 w-4" />
          Apply
        </button>
        <div className="relative flex-1 max-w-xs ml-auto">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-500" />
          <input
            type="text"
            placeholder="Search messages…"
            value={search}
            onChange={e => setSearch(e.target.value)}
            className="pl-9 pr-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-sm text-white placeholder-gray-500 focus:outline-none focus:border-blue-500 w-full"
          />
        </div>
        <label className="flex items-center gap-2 text-xs text-gray-400 cursor-pointer">
          <input
            type="checkbox"
            checked={autoScroll}
            onChange={e => setAutoScroll(e.target.checked)}
            className="rounded"
          />
          Auto-scroll
        </label>
      </div>

      {/* Log output */}
      <div className="bg-gray-950 rounded-xl border border-gray-800 overflow-hidden flex flex-col flex-1 min-h-0">
        <div
          className="overflow-y-auto flex-1 font-mono text-xs leading-relaxed"
          style={{ maxHeight: 'calc(100vh - 18rem)' }}
        >
          {!connected && !entries.length && (
            <div className="p-4 text-gray-500 text-center">
              journalctl not available or connecting…
            </div>
          )}
          {filtered.map((e, i) => {
            const p = PRIORITY_LABELS[e.priority] ?? PRIORITY_LABELS[7]
            return (
              <div
                key={i}
                className="flex gap-3 px-4 py-0.5 hover:bg-gray-900/50 border-b border-gray-900"
              >
                <span className="text-gray-600 shrink-0 w-20">{fmtTime(e.timestamp)}</span>
                <span className={`shrink-0 w-12 ${p.cls}`}>{p.label}</span>
                <span className="text-purple-400 shrink-0 w-40 truncate">{e.unit}</span>
                <span className="text-gray-300 break-all">{e.message}</span>
              </div>
            )
          })}
          <div ref={bottomRef} />
        </div>
        <div className="border-t border-gray-800 px-4 py-2 text-xs text-gray-600">
          {filtered.length} entries{search ? ` (filtered from ${entries.length})` : ''}
          {entries.length >= MAX_ENTRIES && ` · capped at ${MAX_ENTRIES}`}
        </div>
      </div>
    </div>
  )
}
