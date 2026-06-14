import { useEffect, useRef, useState } from 'react'

export interface MetricsSnapshot {
  ts: string
  cpu_pct: number
  mem_total_kb: number
  mem_free_kb: number
  mem_used_kb: number
  swap_total_kb: number
  swap_used_kb: number
  load1: number
  load5: number
  load15: number
  uptime_seconds: number
  net_rx_bytes: number
  net_tx_bytes: number
}

export function useMetrics() {
  const [latest, setLatest] = useState<MetricsSnapshot | null>(null)
  const [history, setHistory] = useState<MetricsSnapshot[]>([])
  const wsRef = useRef<WebSocket | null>(null)

  useEffect(() => {
    const proto = window.location.protocol === 'https:' ? 'wss' : 'ws'
    const url = `${proto}://${window.location.host}/ws/metrics`

    let reconnectTimer: ReturnType<typeof setTimeout>

    function connect() {
      const ws = new WebSocket(url)
      wsRef.current = ws

      ws.onmessage = (e) => {
        try {
          const env = JSON.parse(e.data)
          if (env.type === 'metrics') {
            const snap: MetricsSnapshot = env.payload
            setLatest(snap)
            setHistory((h) => [...h.slice(-59), snap])
          }
        } catch { /* ignore parse errors */ }
      }

      ws.onclose = () => {
        reconnectTimer = setTimeout(connect, 3000)
      }
    }

    connect()
    return () => {
      clearTimeout(reconnectTimer)
      wsRef.current?.close()
    }
  }, [])

  return { latest, history }
}
