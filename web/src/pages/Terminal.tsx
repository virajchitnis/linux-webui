import { useEffect, useRef, useState, useCallback } from 'react'
import { Terminal as XTerm } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { api } from '@/lib/api'
import '@xterm/xterm/css/xterm.css'

export default function Terminal() {
  const containerRef = useRef<HTMLDivElement>(null)
  const xtermRef = useRef<XTerm | null>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const fitRef = useRef<FitAddon | null>(null)
  const [status, setStatus] = useState<'idle' | 'connecting' | 'connected' | 'error'>('idle')
  const [errMsg, setErrMsg] = useState('')

  const connect = useCallback(async () => {
    if (wsRef.current) return
    setStatus('connecting')
    setErrMsg('')

    let sessionId: string
    try {
      const res = await api.post<{ id: string }>('/api/terminal/new', {})
      sessionId = res.id
    } catch (e) {
      setStatus('error')
      setErrMsg((e as Error).message)
      return
    }

    const xterm = new XTerm({
      theme: {
        background: '#030712',
        foreground: '#d1d5db',
        cursor: '#60a5fa',
        black: '#1f2937',
        red: '#ef4444',
        green: '#22c55e',
        yellow: '#eab308',
        blue: '#3b82f6',
        magenta: '#a855f7',
        cyan: '#06b6d4',
        white: '#f9fafb',
        brightBlack: '#374151',
        brightRed: '#f87171',
        brightGreen: '#4ade80',
        brightYellow: '#facc15',
        brightBlue: '#60a5fa',
        brightMagenta: '#c084fc',
        brightCyan: '#22d3ee',
        brightWhite: '#ffffff',
      },
      fontFamily: '"Fira Code", "Cascadia Code", "JetBrains Mono", monospace',
      fontSize: 13,
      cursorBlink: true,
      scrollback: 1000,
    })

    const fit = new FitAddon()
    xterm.loadAddon(fit)
    xtermRef.current = xterm
    fitRef.current = fit

    if (containerRef.current) {
      xterm.open(containerRef.current)
      fit.fit()
    }

    const proto = window.location.protocol === 'https:' ? 'wss' : 'ws'
    const ws = new WebSocket(`${proto}://${window.location.host}/ws/terminal/${sessionId}`)
    wsRef.current = ws

    ws.onopen = () => setStatus('connected')
    ws.onmessage = (ev) => {
      const msg = JSON.parse(ev.data as string)
      if (msg.type === 'output') {
        xterm.write(msg.payload.data as string)
      } else if (msg.type === 'error') {
        xterm.writeln(`\r\n\x1b[31mError: ${(msg.payload as { message: string }).message}\x1b[0m`)
        setStatus('error')
      }
    }
    ws.onerror = () => { setStatus('error'); setErrMsg('WebSocket connection failed') }
    ws.onclose = () => {
      setStatus('idle')
      wsRef.current = null
      xterm.writeln('\r\n\x1b[33m[Session closed]\x1b[0m')
    }

    // Forward keyboard input to server.
    xterm.onData(data => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'input', payload: { data } }))
      }
    })

    // Send resize events.
    const sendResize = () => {
      if (ws.readyState === WebSocket.OPEN) {
        const { rows, cols } = xterm
        ws.send(JSON.stringify({ type: 'resize', payload: { rows, cols } }))
      }
    }
    xterm.onResize(sendResize)
    sendResize()
  }, [])

  // Resize xterm on window resize.
  useEffect(() => {
    const onResize = () => fitRef.current?.fit()
    window.addEventListener('resize', onResize)
    return () => window.removeEventListener('resize', onResize)
  }, [])

  // Cleanup on unmount.
  useEffect(() => {
    return () => {
      wsRef.current?.close()
      xtermRef.current?.dispose()
    }
  }, [])

  const disconnect = () => {
    wsRef.current?.close()
    wsRef.current = null
    xtermRef.current?.dispose()
    xtermRef.current = null
    fitRef.current = null
    setStatus('idle')
    // Clear container
    if (containerRef.current) containerRef.current.innerHTML = ''
  }

  return (
    <div className="h-full flex flex-col gap-4">
      <div className="flex items-center justify-between flex-shrink-0">
        <div>
          <h1 className="text-2xl font-bold text-white">Terminal</h1>
          <p className="text-gray-400 text-sm mt-0.5">Shell session</p>
        </div>
        <div className="flex items-center gap-3">
          <StatusDot status={status} />
          {status === 'idle' || status === 'error' ? (
            <button
              onClick={connect}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm rounded-lg transition-colors"
            >
              Open terminal
            </button>
          ) : (
            <button
              onClick={disconnect}
              className="px-4 py-2 bg-gray-800 hover:bg-gray-700 border border-gray-700 text-white text-sm rounded-lg transition-colors"
            >
              Close
            </button>
          )}
        </div>
      </div>

      {errMsg && (
        <div className="bg-red-900/20 border border-red-800 rounded-lg px-4 py-3 text-red-400 text-sm flex-shrink-0">
          {errMsg}
        </div>
      )}

      <div
        className="bg-gray-950 rounded-xl border border-gray-800 overflow-hidden flex-1 min-h-0"
        style={{ minHeight: '400px' }}
      >
        {status === 'idle' && !errMsg && (
          <div className="h-full flex items-center justify-center text-gray-500 text-sm">
            Click "Open terminal" to start a shell session
          </div>
        )}
        <div
          ref={containerRef}
          className="h-full p-2"
          style={{ display: status !== 'idle' ? 'block' : 'none' }}
        />
      </div>
    </div>
  )
}

function StatusDot({ status }: { status: 'idle' | 'connecting' | 'connected' | 'error' }) {
  const colors: Record<typeof status, string> = {
    idle: 'bg-gray-500',
    connecting: 'bg-yellow-400 animate-pulse',
    connected: 'bg-green-400',
    error: 'bg-red-400',
  }
  const labels: Record<typeof status, string> = {
    idle: 'Disconnected',
    connecting: 'Connecting…',
    connected: 'Connected',
    error: 'Error',
  }
  return (
    <div className="flex items-center gap-2 text-xs text-gray-400">
      <div className={`h-2 w-2 rounded-full ${colors[status]}`} />
      {labels[status]}
    </div>
  )
}
